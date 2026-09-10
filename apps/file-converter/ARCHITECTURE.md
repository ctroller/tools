# file-converter — architecture

## Status

Converter/Registry layer (`internal/convert`), HTTP layer (`/formats`,
`/files`, job worker, SSE) and graceful shutdown are all implemented.

Not yet built: the `download` endpoint, the `Problem` `type`/`title` slug
table, and the pixel-dimension cap.

Top-level repo contract (repo layout, routing/StripPrefix, dev environment)
lives in the root `AGENTS.md`; this file covers file-converter's design only.

## Scope (current)

Raster image formats only: **PNG, JPEG, WEBP, AVIF, GIF**, via libvips.

More converters to follow.

## Architecture pattern

**Compile-time registry + interface** — not runtime plugin loading. Go's
`plugin` package (`.so` loading) was explicitly rejected: Linux-only,
toolchain-version-locked, a known footgun. "Plugin" here means: implement
the interface, register it in an `init()`, rebuild the binary. Adding a
format is a new file + rebuild, never a dynamic drop-in.

This is the Strategy pattern behind a registry — equivalently, Ports &
Adapters: `Converter` is the port, each transformer is an adapter over its
own underlying tech (cgo library, subprocess, whatever a future format
needs). The registry doesn't care which.

```go
type MediaType string // canonical MIME type, e.g. "image/webp", "image/avif"

type Options struct {
    Target MediaType
}

type Converter interface {
    Name() string
    SupportedFormats() map[MediaType][]MediaType
    Convert(ctx context.Context, in io.ReadSeeker, out io.Writer, opts Options) error
}
```

A converter can optionally implement `Lifecycle` (`Start`/`Stop`) if it needs
setup or teardown around the registry's `StartAll`/`StopAll` calls. The
registry type-asserts for it — it's not part of `Converter` itself.
`ImageConverter` uses it for `vips.Startup`/`vips.Shutdown`.

Quality (1–100, 0 = format default) isn't in `Options` yet — add it directly
when actually needed, don't build a generic options schema preemptively.

`io.ReadSeeker` (not plain `io.Reader`) is deliberate: some libraries need
random access into the file structure, not just a forward stream (this is
what a future PDF library would need). The HTTP handler buffers the upload
to a temp file once and passes `*os.File` in — satisfies `io.ReadSeeker` for
everyone, costs nothing for transformers that only need `io.Reader`. That
temp file now persists for the job's lifetime (upload → convert → download),
not just one request — see "Cleanup" below.

**Why `MediaType` is a MIME type, not a bare identifier** ("png"/"webp"/...
would work but throws away things MIME types give for free): magic-byte
detection already returns MIME-type strings, so detection output maps
straight into `MediaType` with no translation table. The `Content-Type`
response header is the same value, again with no lookup.

Detection uses `github.com/gabriel-vasile/mimetype` (magic-byte based),
not stdlib `http.DetectContentType` — chosen for its wider format coverage,
including AVIF.

`SupportedFormats()` replaces an earlier `CanHandle(src, tgt)` predicate
design. Each converter states its full source→targets map directly — e.g.
`ImageConverter` builds all pairs among its five image types, excluding
self, once in its constructor. That table backs both
`Lookup` (routing, at convert time) and `Formats` (used by `POST /files`,
see "HTTP API") directly, so the two can never disagree.

**Registry**: one generic `ImageConverter` claims all PNG/JPEG/WEBP/AVIF
pairs — N decoders + M encoders around libvips' shared pixel buffer, not
NxM point-to-point transformers. Only register a specialized converter for
a pair that genuinely can't go through the generic path.

**MediaType table**: single source of truth (`internal/convert/format.go` or
similar) mapping each `MediaType` (MIME type) to its canonical file
extension. That's the only mapping needed — `MediaType` values are used
directly on the wire, so there's no separate slug vocabulary to keep in
sync. Referenced by the output filename extension and nowhere else needs
it. **Not yet built** — `image.go`'s `ImageTypes` var and `internalConvert`'s
switch statement currently duplicate format knowledge independently.

## HTTP API

Asynchronous, job-based. Upload, target selection, and result retrieval are
three separate calls; conversion runs in a background worker pool, not
inline in a request. No accounts — a random, unguessable `handle` is the
only thing standing in for auth, so treat it as a bearer token: anyone
holding it can read status and download the result. Currently a v4 UUID
from Go 1.27's stdlib `uuid` package (`crypto/rand`-backed under the hood),
generated in `JobIntake.Submit` (`task/intake.go`).

### `GET /formats`

Returns the full `Registry.Formats()` matrix (every known source MediaType
mapped to its list of reachable target MediaTypes). Not tied to any upload —
a client can render a static conversion matrix without uploading first.

### `POST /files`

Multipart file upload. Server saves it to a temp file, detects the
**actual** source format via magic-byte sniffing — never trusts the
client's claimed MIME type or file extension — and creates a job record
(status `uploaded`) keyed by `handle`. Returns
`{handle, detectedSource, targets}`, where `targets` is
`Registry.Formats()[detectedSource]`. Detection happens once, here, against
the real bytes — no client-side extension guessing needed downstream.

### `PUT /files/{handle}/convert?target=<media-type>`

`target` is the literal MIME type, e.g. `?target=image/webp` — no slug
layer, same vocabulary as `Content-Type` and the `/files` response.

Validates `(detectedSource, target)` against the registry. On success, sets
status `pending`, enqueues the job on the worker pool, and returns `202`
with a `Location: /files/{handle}` header and no response body. Rejects a
second `convert` call while a job for that `handle` is already
`pending`/`processing` — one active job per handle (enforced by a
compare-and-swap on the job's status from `uploaded` to `pending`).

### `GET /files/{handle}/events`

SSE stream. **Current implementation**: a 1-second ticker polls the job
record and pushes its current status on every tick, so the first message
lands after roughly one second, not immediately on connect, and the client
also receives repeated identical messages between real transitions. The
stream still closes once status reaches `done`/`failed`. Design intent
(sending state immediately on connect, only on transition after that) is
not yet built — revisit if the repeated messages become a problem.

### `GET /files/{handle}`

Same status data as the SSE stream, as one JSON response. Fallback for a
client not using SSE; cheap, since the record already exists.

### `GET /files/{handle}/download` — not yet built

Design intent, not current behavior: stream the converted file when status
is `done`, and delete the output file after a successful transfer — see
"Cleanup". No route for this exists yet in `router.go`; a job can reach
`done` with no way to fetch its result over HTTP.

### Validation & limits

Two independent limits, deliberately living at different layers:

- **Upload size cap: 25MB.** Generic, format-agnostic — currently enforced
  via `r.ParseMultipartForm(25 << 20)` in `Files` (`httpapi/files.go`), not
  `http.MaxBytesReader`. This is a soft cap: the argument to
  `ParseMultipartForm` is the in-memory threshold before a part spills to a
  temp file, not a hard ceiling on total request body size. An oversize
  request currently surfaces as a generic `400`, indistinguishable from any
  other multipart parse error — see "Errors" below. Revisit with
  `http.MaxBytesReader` if a true hard cap and a distinct `413` are needed.
- **Pixel dimension cap: 12000×12000px.** MediaType-specific — enforced inside
  the libvips transformer via a cheap header peek *before* full decode.
  This exists because file size on disk doesn't bound decoded memory use: a
  tiny file can declare enormous dimensions (decompression-bomb shape) and
  blow memory well past what the byte cap would catch.
  **Not yet built** — `ImageConverter.Convert` decodes immediately with no
  header-only peek first.

Rule for future transformers: **generic limits live in the HTTP layer,
format-specific resource limits live inside the transformer** (a PDF
transformer's equivalent concern would be page count, not pixel dimensions
— it shouldn't need the HTTP layer to know that).

Concurrent work is capped by the worker pool rather than by request/response
backpressure, since jobs no longer run inline in a request.
**Current implementation**: hardcoded to a queue buffer of 5 and exactly 1
worker (`task.NewQueue(5, 1, ...)` in `cmd/server/main.go`), not scaled to
CPU count. Revisit if throughput becomes a bottleneck.

### Cleanup

**Current implementation**: only the output file is deleted, and only when
conversion fails (`JobExecutor.Run` in `task/executor.go`). The input file
is never explicitly deleted after a job finishes, success or failure — it
relies entirely on the periodic sweep below. A `FileJanitor`
(`task/filestore.go`) runs every 15 minutes and deletes any file in the
upload directory older than a 5-minute TTL, regardless of status (both
values hardcoded in `cmd/server/main.go`). This catches an upload that
never got a target chosen, a finished job's leftover input file, and (once
built) a result never downloaded.

Design intent, not yet reachable: delete the output file right after a
successful download — moot until the `download` endpoint exists.

### Errors

Two shapes, because failures now happen at two different times:

**Request-time** (bad `target`, unknown `handle`, size cap exceeded,
unrecognized upload format, job already in progress) —
`application/problem+json` (RFC 7807):

```go
type Problem struct {
    Type     string `json:"type"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"` // reserved — see request-id note below
    Data     any    `json:"data,omitempty"`
}
```

Design intent — the table below is the target shape:

| Case                                         | Status | `type`                    | `title`                  |
|-----------------------------------------------|--------|---------------------------|--------------------------|
| Missing/invalid `target` param               | 400    | `invalid-request`         | Invalid request          |
| Unknown `handle`                             | 404    | `not-found`                | Not found                |
| No transformer for `(detected, target)`      | 400    | `unsupported-conversion`  | Unsupported conversion   |
| Job already in progress for `handle`         | 409    | `conversion-in-progress`  | Conversion in progress   |
| Upload exceeds size cap                      | 413    | `payload-too-large`       | Payload too large        |
| Image dimensions exceed cap                  | 413    | `image-too-large`         | Image too large          |
| Content doesn't match any known input format | 415    | `unrecognized-format`     | Unrecognized file format |

`type` values are meant to be stable slugs, not real dereferenceable URLs —
fine per spec, no docs site needed for a personal project.

**Current implementation gap**: none of the `type`/`title` slugs above are
wired up. Every call to `HttpProblem` passes an empty `type`, which renders
as `type: "about:blank"`, and titles are generic HTTP reason phrases
(`"Bad Request"`, `"Not Found"`, `"Conflict"`, `"Unsupported media type"`),
not the domain-specific titles in the table. Two status codes also differ
from the table:

- No transformer for `(detected, target)` returns **404**, not 400 — it
  reuses the same `common.NotFoundErr` path as an unknown `handle`
  (`task/intake.go`, `httpapi/convert.go`), so the two cases are currently
  indistinguishable to a client.
- Upload exceeds size cap returns a generic **400**, not 413 — see
  "Validation & limits" above.

`instance` is reserved for an **app-generated request ID**, attached by a
logging middleware. **Not built yet** — the field is always empty.

**Job-time** (the worker's `Convert` call fails on otherwise-valid input):
the enqueuing request already returned `202`, so this can't be an HTTP
error response. Decided: this does **not** reuse the `Problem` shape.
Surfaced as status `failed` plus a plain error string
(`JobResult.Error()`), delivered via the SSE stream / `GET /files/{handle}`
as `FileHandleResult.Error`.

## Response contract (`GET /files/{handle}/download`) — not yet built

Design intent for when this endpoint is added:

- **Filename**: strip the original filename's extension, append the
  canonical extension for the target format (from the format table). Fall
  back to `converted.<ext>` if the original filename is missing/empty after
  sanitization.
- Sanitize the client-supplied filename before it touches any header — strip
  path separators and control characters.
- `Content-Disposition: attachment; filename="<ascii-safe>"; filename*=UTF-8''<percent-encoded>`
  per RFC 6266 — ASCII fallback plus a UTF-8 variant for non-ASCII original
  filenames.
- `Content-Type`: the job's `target` value — already the MIME type.
- The output file is complete on disk by the time `download` is reachable
  (status `done` implies the worker finished writing it), so the response
  can stream straight from disk. Unlike the old synchronous design, no
  buffer-then-write trick is needed to guarantee an atomic response — job
  status already guarantees completeness before download is ever callable.

## Container / deploy

cgo (govips for libvips) means **not a static binary** — the runtime
image needs libvips' shared libs present. Multi-stage Dockerfile; runtime
base is **Debian-slim, not Alpine** (musl + cgo + libvips is a known
pain point). Fits the k3s single-node 8GB budget fine, but won't be as lean
as a pure static-Go image.

## Frontend contract (file-converter specific)

Backend is a **pure JSON API + binary streaming** — no HTML fragments, ever,
regardless of what the platform-wide frontend ends up looking like. This is
why htmx-style server-rendered responses don't fit this API.

Expected flow: drop file → `POST /files` → show target buttons from the
returned `targets` (accurate, since detection already happened server-side
against the real bytes) → `PUT /files/{handle}/convert?target=...` → open
`GET /files/{handle}/events` (SSE) → on `done`, fetch
`GET /files/{handle}/download`. The last step is not yet reachable — see
"Not yet built" above.

## Open / not yet decided

- **Download endpoint** — not built; a job can reach `done` with no way to
  fetch the result over HTTP.
- **`Problem` `type`/`title` slugs** — not wired up; every error currently
  renders `type: "about:blank"` with a generic title.
- **Two error status codes don't match the design table** — unsupported
  `(detected, target)` returns 404 instead of 400, and an oversize upload
  returns 400 instead of 413. Fix the code or update the table, whichever
  turns out to be the right target behavior.
- **Upload size cap mechanism** — `ParseMultipartForm`'s argument is a
  memory threshold, not a hard body-size ceiling. Revisit with
  `http.MaxBytesReader` for a true 25MB cap.
- **Input-file cleanup** — no explicit delete after a job finishes; relies
  entirely on the periodic TTL sweep.
- **SSE cadence** — ticker-based polling sends repeated identical messages
  and delays the first message by ~1 second, instead of pushing
  immediately on connect and only on transitions after that.
- Testing approach — per-package unit tests exist
  (`internal/convert/*_test.go`), but no decision on fixture-based table
  tests per transformer as coverage grows.
- Logging level/format — `log/slog` is in use throughout, but no
  conventions are set for levels or structured fields.
- Request-ID middleware — needed to actually populate `Problem.instance`.
- Worker pool sizing — hardcoded to 1 worker / buffer 5, not CPU-scaled.
- Pixel dimension cap — not built.
- No liveness/readiness endpoint decided yet, despite deploying to k3s.
