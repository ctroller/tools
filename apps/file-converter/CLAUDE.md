# file-converter

## Status

No code yet. This document is the settled architecture from a design
discussion (2026-08-18) — read it before scaffolding, it's the re-entry
point for this app specifically. Top-level repo contract (repo layout,
routing/StripPrefix, dev environment) lives in the root `CLAUDE.md`; this
file only covers what's specific to file-converter.

## Scope (current)

Raster image formats only: **PNG, JPEG, WEBP, AVIF**, via libvips.

**PDF is explicitly deferred** — not being built now, but it was used as a
stress test while designing the interface below, which is why the interface
already accommodates it without a redesign. See "Future: PDF" at the bottom
for how it's expected to slot in later.

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

type ConvertOptions struct {
    Quality int // 1–100; 0 = format default. Not exposed via the API in v1 —
                // add fields here directly when actually needed, don't build
                // a generic options schema/config system preemptively.
}

type Converter interface {
    CanHandle(src, tgt MediaType) bool
    Convert(ctx context.Context, in io.ReadSeeker, out io.Writer, opts ConvertOptions) error
}
```

`io.ReadSeeker` (not plain `io.Reader`) is deliberate: some libraries need
random access into the file structure, not just a forward stream (this is
what a future PDF library would need). The HTTP handler buffers the upload
to a temp file once and passes `*os.File` in — satisfies `io.ReadSeeker` for
everyone, costs nothing for transformers that only need `io.Reader`.

**Why `MediaType` is a MIME type, not a bare identifier** ("png"/"webp"/...
would work but throws away things MIME types give for free): magic-byte
detection (`http.DetectContentType`, or libvips' own format identification
as a fallback if stdlib sniffing doesn't cover a given format — check AVIF
coverage for the target Go version, WEBP has been supported a long time)
already returns MIME-type strings, so detection output maps straight into
`MediaType` with no translation table. The `Content-Type` response header is
the same value, again with no lookup. And it gives `CanHandle` a namespace
to pattern-match on:

```go
func (t *libvipsTransformer) CanHandle(src, tgt MediaType) bool {
    return strings.HasPrefix(string(src), "image/") &&
           strings.HasPrefix(string(tgt), "image/") &&
           src != tgt
}
```
That's "any image to any other image" in three lines — no interface change
needed to support it, since `CanHandle` is a predicate, not a static list.
The same pattern would cover "any file" if a transformer ever genuinely
needed that breadth, but nothing today does — don't build for it until
something concrete does.

**Registry**: one generic `libvipsTransformer` claims all PNG/JPEG/WEBP/AVIF
pairs — N decoders + M encoders around libvips' shared pixel buffer, not
NxM point-to-point transformers. Only register a specialized transformer for
a pair that genuinely can't go through the generic path.

**MediaType table**: single source of truth (`internal/convert/format.go` or
similar) mapping each `MediaType` (MIME type) to its canonical file
extension. That's the only mapping needed — `MediaType` values are used
directly on the wire (see HTTP API below), so there's no separate slug
vocabulary to keep in sync. Referenced by the output filename extension and
nowhere else needs it, since `Content-Type` and `/formats`/`?target=` are
already the MIME type itself.

## HTTP API

Synchronous only. No accounts, no job queue, no websocket/polling — libvips
converts in milliseconds at this scale, so a blocking request/response is
the whole feature.

### `GET /formats`

Derived from the registry (not hand-maintained), returns the supported
conversion matrix keyed and valued by MIME type directly, e.g.
`{"image/png": ["image/webp","image/avif","image/jpeg"], ...}` — no slug
vocabulary, same values used everywhere else (detection, `target`,
`Content-Type`). Frontend fetches this once instead of hardcoding the
matrix — single source of truth as the registry grows.

Built by enumerating the **closed set of formats in the MediaType table**
and testing each pair against `CanHandle` — not by asking transformers to
enumerate everything they could theoretically handle. That distinction
matters now that `CanHandle` can match an open-ended space (e.g. an
`image/*` prefix check): discovery is driven by what you choose to
advertise, not by introspecting the transformer's matching logic.

### `POST /convert?target=<media-type>`

`target` is the literal MIME type, e.g. `?target=image/webp` — no slug
layer. A `/` doesn't need percent-encoding inside a query string
(RFC 3986: `query = *( pchar / "/" / "?" )`), so there's no URL-ugliness
reason to translate it to something shorter, and using the same vocabulary
everywhere means one less mapping to maintain and one less place a lookup
can go stale.

Multipart file upload in. Server detects the **actual** source format via
magic-byte sniffing — never trusts the client's claimed MIME type or file
extension. Looks up `(detected, target)` in the registry. On success,
buffers the full converted output in memory before writing anything to the
response (see "Response contract" below for why), then writes it.

No separate `/detect` endpoint — detection happens inline as part of
`/convert`. Two round trips would buy nothing here.

### Validation & limits

Two independent limits, deliberately living at different layers:

- **Upload size cap: 25MB.** Generic, format-agnostic — enforced via
  `http.MaxBytesReader` in the HTTP layer, before the body is read.
- **Pixel dimension cap: 12000×12000px.** MediaType-specific — enforced inside
  the libvips transformer via a cheap header peek *before* full decode.
  This exists because file size on disk doesn't bound decoded memory use: a
  tiny file can declare enormous dimensions (decompression-bomb shape) and
  blow memory well past what the byte cap would catch. ~500MB–1GB worst
  case per request at this cap, tolerable on the single-node 8GB box.

Rule for future transformers: **generic limits live in the HTTP layer,
format-specific resource limits live inside the transformer** (a PDF
transformer's equivalent concern would be page count, not pixel dimensions
— it shouldn't need the HTTP layer to know that).

Explicitly not built: concurrent-request limiting (a semaphore capping
simultaneous conversions). Flagged as the natural next knob if memory
pressure becomes a real, observed problem — not built preemptively.

### Errors — `application/problem+json` (RFC 7807)

```go
type Problem struct {
    Type     string `json:"type"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"` // reserved — see request-id note below
}
```

| Case | Status | `type` | `title` |
|---|---|---|---|
| Missing/invalid `target` param | 400 | `invalid-request` | Invalid request |
| No transformer for `(detected, target)` | 400 | `unsupported-conversion` | Unsupported conversion |
| Upload exceeds size cap | 413 | `payload-too-large` | Payload too large |
| Image dimensions exceed cap | 413 | `image-too-large` | Image too large |
| Content doesn't match any known input format | 415 | `unrecognized-format` | Unrecognized file format |
| Transformer failure on otherwise-valid input | 500 | `about:blank` | Internal error |

`type` values are stable slugs, not real dereferenceable URLs — fine per
spec, no docs site needed for a personal project. 500 uses `about:blank`
deliberately (the spec's documented default) since `detail` must never leak
internal error strings.

`instance` is reserved for an **app-generated request ID** (random, e.g.
`crypto/rand`-based — no UUID dependency needed), attached by a logging
middleware. **Not built yet.** Explicitly out of scope for now: real
distributed tracing (W3C `traceparent`/OpenTelemetry) — that needs an actual
tracing backend (Jaeger/Tempo) which is separate future "observability
stack" work at the platform level, not this app. If that ever gets built,
Traefik has native OTel tracing support and the swap is natural — not a
redesign.

## Response contract (success)

- **Filename**: strip the client-supplied original filename's extension,
  append the canonical extension for the target format (from the format
  table). Fall back to `converted.<ext>` if the original filename is
  missing/empty after sanitization.
- Sanitize the client-supplied filename before it touches any header — strip
  path separators and control characters. Not a filesystem risk (never
  written to disk under that name), but a real header-injection-adjacent
  correctness issue if interpolated raw.
- `Content-Disposition: attachment; filename="<ascii-safe>"; filename*=UTF-8''<percent-encoded>`
  per RFC 6266 — ASCII fallback plus a UTF-8 variant for non-ASCII original
  filenames.
- `Content-Type`: exactly the `target` value from the request — it's already
  the MIME type, no lookup needed.
- `Content-Length`: set explicitly, which means **the handler buffers the
  full converted output in memory before writing anything to the
  `ResponseWriter`**, rather than streaming conversion output as it's
  produced. Deliberate: if conversion fails partway through a true stream,
  a `200` and partial bytes may already be sent, making a clean
  `problem+json` error response impossible. Buffering first guarantees a
  conversion failure is always caught before any header or byte reaches the
  client.

## Container / deploy

cgo (govips/bimg for libvips) means **not a static binary** — the runtime
image needs libvips' shared libs present. Multi-stage Dockerfile; runtime
base should be **Debian-slim, not Alpine** (musl + cgo + libvips is a known
pain point). Fits the k3s single-node 8GB budget fine, but won't be as lean
as a pure static-Go image — set that expectation going in.

## Frontend contract (file-converter specific)

Backend is a **pure JSON API + binary streaming** — no HTML fragments, ever,
regardless of what the platform-wide frontend ends up looking like (that's
a separate discussion, see `.claude/history/frontend-notes.md`). This is
why htmx-style server-rendered responses don't fit this API.

Expected flow: drop file → client-side extension guess narrows the target
buttons shown, from the fetched `/formats` matrix → `POST /convert` →
browser downloads via `Content-Disposition`.

## Open / not yet decided

- Testing approach — not discussed beyond "should exist." No decision on
  fixture-based table tests per transformer, etc.
- Logging level/format — not discussed at all beyond the request-ID note
  above.
- Request-ID middleware — needed to actually populate `Problem.instance`.

## Future: PDF (deferred, not building)

If/when this happens:

- Likely a **subprocess-based transformer** (`poppler-utils`' `pdftoppm` via
  `os/exec`) rather than cgo-mupdf bindings or pure-Go `pdfcpu` — `pdfcpu`
  does PDF *manipulation* (merge/split/watermark), not real rasterization.
  Subprocess keeps it dependency-light; the `Convert(ctx, in, out, opts)`
  interface doesn't care whether an implementation is a cgo call or a
  subprocess, which is exactly why this stays a clean addition.
- Multi-page PDFs break the 1-in/1-out interface shape (`io.Writer out`
  assumes one output blob). Resolve by scoping down, not by generalizing
  the interface: rasterize one page at a time via an explicit `page` option
  in `Options`, called once per page if all pages are wanted. Don't
  change `Convert`'s signature to return multiple outputs.
- Adds another runtime dependency to the container image (`poppler-utils`
  package) if it lands.
