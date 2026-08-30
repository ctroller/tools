# file-converter — architecture

## Status

Converter/Registry layer implemented (`internal/convert`). HTTP layer
(`/files`, job worker, SSE) not yet built — `cmd/server/main.go`'s mux has
no routes registered yet.

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
detection (`http.DetectContentType`, or libvips' own format identification
as a fallback if stdlib sniffing doesn't cover a given format — check AVIF
coverage for the target Go version, WEBP has been supported a long time)
already returns MIME-type strings, so detection output maps straight into
`MediaType` with no translation table. The `Content-Type` response header is
the same value, again with no lookup.

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
inline in a request. No accounts — a random, unguessable `handle`
(`crypto/rand`-based) is the only thing standing in for auth, so treat it as
a bearer token: anyone holding it can read status and download the result.

### `POST /files`

Multipart file upload. Server saves it to a temp file, detects the
**actual** source format via magic-byte sniffing — never trusts the
client's claimed MIME type or file extension — and creates a job record
(status `uploaded`) keyed by `handle`. Returns
`{handle, detectedSource, targets}`, where `targets` is
`Registry.Formats()[detectedSource]`. Detection happens once, here, against
the real bytes — no client-side extension guessing needed downstream.

### `POST /files/{handle}/convert?target=<media-type>`

`target` is the literal MIME type, e.g. `?target=image/webp` — no slug
layer, same vocabulary as `Content-Type` and the `/files` response.

Validates `(detectedSource, target)` against the registry. On success, sets
status `queued`, enqueues the job on the worker pool, returns `202` +
`{status: "queued"}`. Rejects a second `convert` call while a job for that
`handle` is already `queued`/`processing` — one active job per handle.

### `GET /files/{handle}/events`

SSE stream. Sends the job's current status immediately on connect (so a
client connecting after processing already finished still gets the right
state), then pushes each subsequent transition (`processing → done | failed`).

### `GET /files/{handle}`

Same status data as the SSE stream, as one JSON response. Fallback for a
client not using SSE; cheap, since the record already exists.

### `GET /files/{handle}/download`

Streams the converted file when status is `done`. Deletes the output file
after a successful transfer — see "Cleanup".

### Validation & limits

Two independent limits, deliberately living at different layers:

- **Upload size cap: 25MB.** Generic, format-agnostic — enforced via
  `http.MaxBytesReader` in the HTTP layer, before the body is read.
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

Concurrent work is capped by the worker pool size (bounded, sized to CPU
count) rather than by request/response backpressure, since jobs no longer
run inline in a request. Pool size: not yet decided — see "Open / not yet
decided".

### Cleanup

Output file deleted right after a successful download. Input file deleted
once conversion finishes, success or failure. A periodic sweep additionally
deletes anything past a TTL regardless of status, catching an upload that
never got a target chosen, or a result never downloaded. TTL value: not yet
decided.

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
}
```

| Case                                         | Status | `type`                    | `title`                  |
|-----------------------------------------------|--------|---------------------------|--------------------------|
| Missing/invalid `target` param               | 400    | `invalid-request`         | Invalid request          |
| Unknown `handle`                             | 404    | `not-found`                | Not found                |
| No transformer for `(detected, target)`      | 400    | `unsupported-conversion`  | Unsupported conversion   |
| Job already in progress for `handle`         | 409    | `conversion-in-progress`  | Conversion in progress   |
| Upload exceeds size cap                      | 413    | `payload-too-large`       | Payload too large        |
| Image dimensions exceed cap                  | 413    | `image-too-large`         | Image too large          |
| Content doesn't match any known input format | 415    | `unrecognized-format`     | Unrecognized file format |

`type` values are stable slugs, not real dereferenceable URLs — fine per
spec, no docs site needed for a personal project.

`instance` is reserved for an **app-generated request ID**, attached by a
logging middleware. **Not built yet.**

**Job-time** (the worker's `Convert` call fails on otherwise-valid input):
the enqueuing request already returned `202`, so this can't be an HTTP
error response. Surfaced as status `failed` + an error message on the job
record, delivered via the SSE stream / `GET /files/{handle}`. Whether that
error message should reuse the `Problem` shape, for consistency: not yet
decided.

## Response contract (`GET /files/{handle}/download`)

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
against the real bytes) → `POST /files/{handle}/convert?target=...` → open
`GET /files/{handle}/events` (SSE) → on `done`, fetch
`GET /files/{handle}/download`.

## Open / not yet decided

- Testing approach — not discussed beyond "should exist." No decision on
  fixture-based table tests per transformer, etc.
- Logging level/format — not discussed at all beyond the request-ID note
  above.
- Request-ID middleware — needed to actually populate `Problem.instance`.
- Worker pool size.
- TTL duration for the cleanup sweep.
- Whether job-time failures should reuse the `Problem` shape.
- No liveness/readiness endpoint decided yet, despite deploying to k3s.
