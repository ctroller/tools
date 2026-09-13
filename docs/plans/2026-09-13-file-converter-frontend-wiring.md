# File converter — frontend wiring: task breakdown

Personal task plan. Not an agent-execution plan — no code solutions here,
by `AGENTS.md`'s rule for this repo. Each task lists its contract (types,
signatures, endpoints), the Svelte/SvelteKit concepts it touches, and what
"done" looks like. Copy each task section into your PM tool / kanban as
one card.

## Goal

Wire the file-converter tool page to the real backend API, for one file
at a time: drop a file, pick a target format, watch the job run, see a
result message. No download step yet.

## Decisions already locked

- **Download endpoint**: out of scope. Backend doesn't have it yet
  either (`ARCHITECTURE.md` marks it "not yet built"). Stop at the
  `done` status with a message, no file transfer.
- **One file at a time**: `FileDrop` currently allows multiple files,
  but the backend takes one file per upload/job. Restrict the UI to a
  single file for this pass.
- **Status updates**: SSE (`GET /files/{handle}/events`) as primary
  channel, falling back to polling `GET /files/{handle}` if the stream
  errors before the job reaches a terminal status.

## Backend contract (already built, verified against source)

- `POST /files` — multipart field name is `file` (singular, not
  `files[]`). Returns `{ data: { handle, detectedSource, targets } }`.
  `targets` is a list of MIME type strings the detected source can
  convert to.
- `POST /files/{handle}/convert?target=<mime>` — **note**: the code is
  `POST`. `ARCHITECTURE.md` says `PUT` — that doc is stale on this one
  point, trust the router source (`apps/file-converter/internal/httpapi/router.go`). Returns `202`,
  empty body.
- `GET /files/{handle}/events` — SSE, one message per second. Payload:
  `{ data: { status, error?, handle } }`.
- `GET /files/{handle}` — same payload, one-shot poll.
- `status` is one of: `uploaded | pending | processing | done | failed`.
- Errors: non-2xx responses are `application/problem+json`:
  `{ type, title, status, detail?, instance? }` (RFC 7807). This is **not** the envelope in
  `src/lib/response-types.ts` — that file
  models a different, unused convention. Don't reuse it here; its
  `success`/`errors` shape doesn't match what this backend sends.

## Libraries

- **Use native `fetch`**, not axios. No dependency to add, handles
  `FormData` and JSON fine, and this API doesn't need interceptors or
  retry logic that would justify axios's weight. If you want automatic
  retries or timeout handling later, `ky` is a smaller, fetch-based
  alternative worth a look before reaching for axios.
- **Use native `EventSource`** for the SSE channel — it's the standard
  browser API for `text/event-stream`, no library needed. Its main
  limitations (no custom headers, no POST) don't apply here — this
  endpoint is a plain unauthenticated `GET`.

## Task A — API client: request/response layer

**Depends on:** nothing (start here). **File:** new `apps/frontend/src/lib/api/file-converter.ts`.

Three thin `fetch` wrappers plus the shared parsing they need:

- `uploadFile(file: File): Promise<UploadResult>` — `POST /files` as
  `FormData` with a `file` field.
- `startConversion(handle: string, target: MediaType): Promise<void>` —
  `POST /files/{handle}/convert?target=...`. Remember to
  `encodeURIComponent` both the handle and the target — the target is a
  MIME type and contains a `/`.
- `fetchJobStatus(handle: string): Promise<JobStatusResult>` —
  `GET /files/{handle}`.
- A shared response parser: on `res.ok`, unwrap `{ data: T }`; on
  failure, parse the `ProblemDetails` body and throw it wrapped in a
  small `FileConverterError extends Error` (carry the parsed
  `ProblemDetails` on it so callers can branch on `status` if they ever
  need to, e.g. 409 = "already converting").

**Concepts to look up:**

- Native `fetch` with `FormData` bodies (don't set `Content-Type`
  yourself — the browser sets the multipart boundary for you).
- `Response.ok` / `res.json()` error-branch pattern.
- TypeScript discriminated envelope typing (`{ data: T }` vs a problem
  shape).

**Testing:** vitest, `vi.stubGlobal('fetch', vi.fn())` per test, assert
the request URL/method/body and the parsed return value or thrown
`FileConverterError`. Cover at least one success and one error case per
function.

**Done when:** all three functions have passing unit tests against a
mocked `fetch`, for both success and a representative error response (e.g. 415 on upload, 409 on convert, 404 on
status).

## Task B — API client: live job updates

**Depends on:** Task A (reuses `fetchJobStatus` and the envelope type). **File:** same file,
`apps/frontend/src/lib/api/file-converter.ts`.

- `subscribeToJob(handle: string, onUpdate: (u: JobStatusResult) => void): () => void`.
- Opens `new EventSource(...)`, parses each message as
  `{ data: JobStatusResult }`, calls `onUpdate`.
- On `source.onerror`, close the source and start polling
  `fetchJobStatus` on an interval (e.g. 2s) instead.
- Whichever channel reports `done` or `failed` should stop itself (close the SSE source / clear the poll interval).
- Return value is an unsubscribe function that tears down whichever
  channel is currently active — the caller (Task D) needs this for
  cleanup.

**Concepts to look up:**

- `EventSource` — `onmessage`, `onerror`, `.close()`. It auto-reconnects
  on some errors by default; you're intentionally overriding that by
  closing and switching to polling instead.
- The "subscribe returns an unsubscribe closure" pattern — same shape
  Svelte's own stores use, worth recognizing.
- `vi.useFakeTimers()` / `vi.advanceTimersByTimeAsync()` for testing the
  polling branch without a real 2-second wait.

**Testing:** stub `globalThis.EventSource` with a small fake class you
control from the test (expose `emitMessage`/`emitError` methods on it).
Cover: forwarding SSE messages, switching to polling on `onerror`,
stopping both channels once a terminal status arrives, and the
unsubscribe function actually stopping everything.

**Done when:** all three behaviors above have a passing test.

## Task C — `FileDrop.svelte`: single-file rework

**Depends on:** nothing (pure UI change) — but Tasks D and E build on its
new single-file `ondrop` signature, so land this first if you want to
work top-down instead of bottom-up. **File:** `apps/frontend/src/lib/components/FileDrop.svelte` (existing).

- Change the `ondrop` prop type from `(files: File[]) => void` to
  `(file: File) => void`.
- Drop the `multiple` attribute and the `files[]` input name (use
  `file`, singular — matches the backend field name from Task A).
- The file `<input>` currently has no `onchange` handler — the "Choose
  a file" button is a no-op. Wire it to the same single-file selection
  path the drop handler uses.
- The submit `<button>` will now sit on a form whose inputs actually do
  something. Add an `onsubmit` handler that prevents the default
  full-page form submission (there was never a real submit path; this
  closes a latent regression rather than introducing one).

**Concepts to look up:**

- Svelte 5 runes: `$props()`, `$state()`. This component already uses
  them — skim it before editing.
- Controlled file inputs: reading `input.files` in `onchange` vs
  `event.dataTransfer.files` in `ondrop` — same underlying `FileList`,
  two different events.

**Testing:** `@testing-library/svelte`. Cover: choosing a file via the
input calls `ondrop` with that file; dropping multiple files calls
`ondrop` exactly once, with only the first file.

**Done when:** both the picker and the drop zone produce exactly one
`File` via `ondrop`, with tests for each path.

## Task D — `ConversionJob.svelte`: job lifecycle component

**Depends on:** Task A (upload/convert calls), Task B (`subscribeToJob`). **File:** new
`apps/frontend/src/lib/components/ConversionJob.svelte`.

Takes one `{ file: File }` prop and drives the whole lifecycle for it:

- States: `uploading → choosing-target → converting → done | failed`.
  Model as a plain `$state` string union — no state library needed for
  five states.
- On mount, call `uploadFile`. On success, store `handle` and `targets`
  and move to `choosing-target`. On failure, move to `failed` with the
  error message.
- Render one button per entry in `targets`; clicking one calls
  `startConversion(handle, target)`, moves to `converting`, then calls
  `subscribeToJob` and reacts to `done`/`failed` updates.
- On unmount (or if the component is ever given a new `file`), call the
  unsubscribe function from `subscribeToJob` — don't leak the SSE
  connection or poll timer.
- Target format labels: deriving `WEBP` from `image/webp` (split on
  `/`, uppercase) is enough for a first pass — no separate label table
  needed.

**Concepts to look up:**

- Svelte 5 `$effect` — the idiomatic place for "run once when this
  component appears, clean up when it goes away" (its return value is
  the cleanup function, called on unmount or before the effect reruns).
- Component-local state machines as a single `$state` union rather than
  a bunch of independent booleans — avoids impossible combinations.

**Testing:** mock the Task A/B module with `vi.mock('$lib/api/file-converter', ...)`
so you control when upload resolves, when target buttons appear, and
what `subscribeToJob`'s callback receives. Cover: target buttons appear
after a successful upload, the upload-error message renders, clicking a
target starts conversion and reacts to a `done` update, and reacts to a
`failed` update with the job's error text.

**Done when:** the full state path (upload → choose target → convert →
done/failed) is exercised by tests without touching a real network.

## Task E — Wire the tool page

**Depends on:** Task C (new `FileDrop` signature), Task D (`ConversionJob`). **File:**
`apps/frontend/src/routes/tools/file-converter/+page.svelte`
(existing — currently just logs the dropped files).

- Hold the dropped `File` in page-level `$state`.
- Once a file exists, render a `ConversionJob` for it inside the
  existing `ToolShell`/`FileDrop` markup.
- Wrap it in a `{#key file}...{/key}` block so dropping a second file (after the first finishes) mounts a fresh
  `ConversionJob` instead of
  reusing the old one's state.

**Concepts to look up:**

- SvelteKit's `{#key}` block — forces a subtree to be torn down and
  rebuilt when the key expression changes; the natural fit for
  "restart this component's whole lifecycle for a new input."
- This repo's test-file naming convention: existing tests for route
  files are named without the leading `+` (`page.test.ts`, not
  `+page.test.ts`) — SvelteKit treats any `+page.*` filename as a route
  module, so a test file needs to avoid that pattern.

**Known risk to check while implementing:** this page renders through
`ToolShell`, which reads `page.url.pathname` from `$app/state` to look
up the current tool. No existing test renders anything through
`ToolShell` yet — the one precedent for touching `$app/state` in a test
is `routes/error.test.ts`, which mocks the whole module (`vi.mock('$app/state', () => ({ page: {...} }))`). You'll
likely need
the same treatment here (mock `page.url.pathname` to
`/tools/file-converter`) — confirm by trying the unmocked render first;
if `ToolShell`'s non-null assertion throws, that's why.

**Testing:** render the page, simulate a file drop/selection, assert a
`ConversionJob`-shaped element appears. You don't need to re-test
`ConversionJob`'s internals here — just that the page hands it the
dropped file.

**Done when:** dropping a file on the real page (via `docker compose up
-d` and the browser) walks through upload → target choice → conversion
→ done/failed against the real `file-converter` container.

## Out of scope (for this pass)

- `GET /files/{handle}/download` — backend endpoint doesn't exist;
  revisit once it does.
- Multi-file drop / batch conversion.
- Fixing the `PUT` vs `POST` mismatch in `ARCHITECTURE.md` — worth a
  one-line doc fix at some point, unrelated to the frontend work here.
- `src/lib/response-types.ts` — unused, mismatched envelope. Not
  touched by this plan; consider deleting it separately if nothing else
  is using it.
