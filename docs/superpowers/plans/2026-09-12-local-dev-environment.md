# Local Dev Environment Implementation Plan

> **For the implementer:** This plan is written for a task/checklist workflow,
> not code-generation. Per this repo's `AGENTS.md`, Claude acts as architect,
> reviewer, and tester here — the repo owner writes the actual Dockerfile,
> config, and script content. Each task below states exactly what must exist
> and how to verify it, without literal file contents. Use
> `superpowers:receiving-code-review` / `superpowers:requesting-code-review`
> for the review checkpoint at the end of each task, on request.

**Goal:** Give every tool a `dev` and `prod` Dockerfile stage, orchestrated
by Compose for dev, with live reload, debugger attach, and editor tooling (via mise) all working — replacing the
devcontainer plan. Except for the
frontend (see Global Constraints), both stages build from one pinned base
image.

**Architecture:** Each `apps/<tool>/Dockerfile` gets a `dev` stage (hot
reload + debug port, source bind-mounted at runtime) alongside its existing
`prod` stage. For most tools, both stages build from the same exact base
image tag; the frontend is an exception (see Global Constraints).
`docker-compose.yaml` builds every tool's `dev` stage and wires them onto
one Compose network.
`mise.toml` pins the same exact toolchain versions on WSL directly, for
editor tooling only — nothing in `mise.toml` runs a dev server.

**Tech Stack:** Docker / Docker Compose, mise, Go + Delve + air (file-converter),
Bun + Vite (frontend).

**Spec:** `docs/superpowers/specs/2026-09-12-local-dev-environment-design.md`

## Global Constraints

- Services must bind `0.0.0.0`, not `127.0.0.1` (`AGENTS.md`'s watch-out —
  directly load-bearing for every dev server this plan adds).
- No Traefik locally — unchanged, prod-only.
- The editor never enters a container. No Dev Containers extension, no
  remote-attach-into-container flow — plain VS Code Remote-WSL throughout.
- A tool's `dev` and `prod` Dockerfile stages must share one exact, pinned
  base image tag. No floating tags (`golang:1.27-trixie`, `oven/bun:1`) once
  this plan is done.
- **Exception: the frontend.** Its `prod` stage runs nginx, not Bun, so it
  cannot share a base image with the Bun-based `dev` stage. The frontend's
  `dev` stage still pins an exact Bun tag. Its `prod` stage keeps the
  floating `nginx:alpine` tag — pinning nginx is out of scope for this plan.

---

### Task 1: `mise.toml` — pin exact toolchain versions

**Files:**

- Create: `mise.toml`

**Interfaces:**

- Produces: the exact Go and Bun version strings that Tasks 2 and
  3 must reuse verbatim as their Dockerfile base image tags.

- [x] **Step 1: Pin Go.** `apps/file-converter/go.mod:3` requires `go 1.27`.
  Find the latest published Go 1.27.x patch release and record its exact
  version (e.g. `1.27.2`).
- [x] **Step 2: Pin Bun.** No version is pinned anywhere today —
  `apps/frontend/package.json` has no `packageManager`/`engines` field, and
  `apps/frontend/Dockerfile:1` floats on `oven/bun:1`. Pick the current
  stable Bun 1.x release and record its exact version.
- [x] **Step 3: Decide whether Node needs a pin at all.** Grep
  `apps/frontend/package.json`'s scripts and any other tooling for a direct
  `node` invocation, as opposed to `bun run ...`. If nothing calls `node`
  directly, do not pin it — Bun covers script execution on its own. If
  something does, record the version it needs.
- [x] **Step 4: Write `mise.toml`** at the repo root, pinning Go and Bun (and Node only if Step 3 found a real need) to
  the exact versions
  recorded above. Not all of these may be first-party mise plugins — check
  mise's own docs for the correct syntax for Go and Bun specifically before
  assuming one uniform `[tools]` entry shape works for both of them.
- [x] **Step 5: Verify install.** Run `mise install` from the repo root.
  Expected: no errors; `mise ls` shows every pinned tool installed at the
  recorded version.
- [x] **Step 6: Verify versions.** Run `mise exec -- go version` and
  `mise exec -- bun --version`.
  Expected: each prints exactly the version recorded in Steps 1-2.
- [x] **Step 7: Commit.**
  ```bash
  git add mise.toml
  git commit -m "chore: Pin toolchain versions for editor tooling via mise"
  ```

---

### Task 2: file-converter Dockerfile — add a `dev` stage

**Files:**

- Modify: `apps/file-converter/Dockerfile`
- Create: `apps/file-converter/.air.toml`

**Interfaces:**

- Consumes: the exact Go version pinned in Task 1.
- Produces: the `dev` stage pattern (stage named `dev`, app port `8080`,
  debug port `2345`) that Task 3 mirrors for the frontend, and that Task 4's
  Compose file builds and publishes.

- [x] **Step 1: Pin the exact base image.** Replace `golang:1.27-trixie`
  (`apps/file-converter/Dockerfile:1`) with the exact patch version recorded
  in Task 1 (e.g. `golang:1.27.2-trixie`), so `dev` and the existing `prod`
  stages share one exact tag.
- [x] **Step 2: Add the `dev` stage.** Base it on that same pinned image.
  It must: install the same `libvips-dev`/`pkg-config` packages the existing
  `builder` stage installs (CGO needs them to compile); install `air` (Go
  hot-rebuild tool) and `dlv` (Delve, the Go debugger), both at pinned
  versions, not `@latest`. It must **not** `COPY . .` — dev source arrives
  via a bind mount at container-run time, not at build time.
- [x] **Step 3: Write `apps/file-converter/.air.toml`.** It must configure:
  a build command that compiles with `-gcflags="all=-N -l"` (disables
  optimizations and inlining, so Delve's line numbers and variable
  inspection stay accurate) instead of the Makefile's plain `go build`; a
  run command that launches the compiled binary through Delve rather than
  running it directly — `dlv exec <path-to-binary> --headless
  --listen=:2345 --api-version=2 --accept-multiclient`.
- [x] **Step 4: Expose the debug port.** Add `EXPOSE 2345` to the `dev`
  stage (the existing `EXPOSE 8080` already covers the HTTP port and stays
  as-is).
- [x] **Step 5: Verify prod is unaffected.**
  ```bash
  docker build -t file-converter:prod-check apps/file-converter
  ```
  Expected: succeeds, same as before this task.
- [x] **Step 6: Verify the dev stage builds.**
  ```bash
  docker build --target dev -t file-converter:dev-check apps/file-converter
  ```
  Expected: succeeds.
- [x] **Step 7: Verify live reload.** Run the dev image with
  `apps/file-converter` bind-mounted to the stage's `WORKDIR` and both ports
  published (`-p 8080:8080 -p 2345:2345`). Confirm a real route from
  `internal/httpapi/router.go` responds via `curl`. Edit a source file,
  save, and confirm the container logs show `air` detecting the change and
  rebuilding, without the container itself restarting.
- [x] **Step 8: Verify debugging.** With the dev container still running,
  connect a debugger to `localhost:2345` (VS Code's Go extension "Connect to
  server" launch config, or `dlv connect localhost:2345`). Set a breakpoint
  in a handler under `internal/httpapi/`, trigger it with a request, and
  confirm execution actually stops there.
- [x] **Step 9: Commit.**
  ```bash
  git add apps/file-converter/Dockerfile apps/file-converter/.air.toml
  git commit -m "feat: Add dev stage to file-converter Dockerfile (air + delve)"
  ```

---

### Task 3: frontend Dockerfile — add a `dev` stage; `prod` stays nginx-based

**Files:**

- Modify: `apps/frontend/Dockerfile`

**Interfaces:**

- Consumes: the exact Bun version pinned in Task 1, for the `dev` stage
  only. The `prod` stage keeps its existing, unpinned `nginx:alpine` tag —
  see Global Constraints' frontend exception.
- Produces: the `dev` stage Task 4's Compose file builds and publishes. The
  existing nginx-based `prod` stage is unchanged; Task 5's docs must
  continue to describe it as nginx-based.

- [x] **Step 1: Pin the exact base image.** Replace the floating `oven/bun:1`
  (`apps/frontend/Dockerfile:1`) with the exact version recorded in Task 1.
- [x] **Step 2: Confirm the build stage's name.** It is already
  `AS build` — no change needed there.
- [x] **Step 3: Add the `dev` stage.** Base it on the same pinned Bun image.
  It must: copy `package.json`/`bun.lock` and run `bun install
  --frozen-lockfile`, matching the `build` stage's own install step; **not**
  copy the rest of the source (bind-mounted at run time instead); run the
  dev server bound to all interfaces (`vite dev`'s `--host` flag needs to be
  `0.0.0.0` explicitly — its default is not that) with Bun's inspector
  enabled on an explicit, fixed port. Check `bun --help` (or Bun's own docs)
  for the exact current flag to run a script while also exposing the
  inspector — don't guess a remembered flag, verify it against this Bun
  version.
- [x] **Step 4: Expose the dev stage's ports.** `EXPOSE` the Vite dev
  server's port (confirm the actual value against `apps/frontend/vite.config.ts`
  — it does not currently override Vite's default) and the explicit
  inspector port chosen in Step 3.
- [x] **Step 5: Verify prod is unaffected.**
  ```bash
  docker build --target prod -t frontend:prod-check apps/frontend
  ```
  Expected: succeeds, same nginx-based image as before this task. `nginx.conf`
  and its `try_files $uri $uri/ /index.html` SPA-fallback rule need no
  change.
- [x] **Step 6: Verify the dev stage builds.**
  ```bash
  docker build --target dev -t frontend:dev-check apps/frontend
  ```
  Expected: succeeds.
- [x] **Step 7: Verify live reload.** Run the dev image bind-mounted with
  both ports published. Confirm the dev server responds over HTTP. Edit a
  `.svelte` file and confirm the change shows up without a manual rebuild.
- [x] **Step 8: Verify debugging.** Connect a debugger to the inspector port
  and confirm it attaches.
- [] **Step 9: Commit.**
  ```bash
  git add apps/frontend/Dockerfile
  git commit -m "feat: Add dev stage to frontend Dockerfile (Vite + Bun inspector)"
  ```

---

### Task 4: `docker-compose.yaml` orchestration

**Files:**

- Modify: `docker-compose.yaml` (currently empty)

**Interfaces:**

- Consumes: file-converter's dev-stage ports (`8080` app, `2345` debug) from
  Task 2; frontend's dev-stage ports from Task 3.
- Produces: `docker compose up -d` as the one command
  that starts every tool's dev container.

- [x] **Step 1: Define the `file-converter` service.** `build.context:
  ./apps/file-converter`, `build.target: dev`; bind-mount
  `./apps/file-converter` into the container's `WORKDIR`; publish the app
  port (`8080`) and the debug port (`2345`) to the host.
- [x] **Step 2: Define the `frontend` service.** `build.context:
  ./apps/frontend`, `build.target: dev`; bind-mount `./apps/frontend` into
  the container's `WORKDIR`; publish the dev-server port and the inspector
  port from Task 3; set `API_PROXY_TARGET=http://file-converter:8080` as an
  environment variable — `apps/frontend/vite.config.ts:16` already reads
  this, no frontend code change needed for it.
- [x] **Step 3: Confirm networking.** Both services need to resolve each
  other by Compose service name. Compose's own default network is normally
  sufficient — only declare a named network explicitly if you have a
  specific reason to.
- [x] **Step 4: Verify it starts.**
  ```bash
  docker compose -f docker-compose.yaml up -d
  docker compose -f docker-compose.yaml ps
  ```
  Expected: both services show as running.
- [x] **Step 5: Verify host access.** `curl` both published app ports from
  the host directly.
- [x] **Step 6: Verify the internal proxy path, before trusting the browser.**
  ```bash
  docker compose -f docker-compose.yaml exec frontend curl http://file-converter:8080/<a-real-route>
  ```
  Expected: succeeds — this proves the Compose-network DNS name resolves
  and `API_PROXY_TARGET` is wired correctly, independent of whether the
  browser-facing proxy happens to look right.
- [x] **Step 7: Commit.**
  ```bash
  git add docker-compose.yaml
  git commit -m "feat: Wire file-converter and frontend dev containers into Compose"
  ```

---

### Task 5: Update `AGENTS.md` and `NOTES.md` to match the finished state

**Files:**

- Modify: `AGENTS.md` (Repo layout, Dev environment sections)
- Modify: `NOTES.md` (Dev experience section) — gitignored, won't show up
  in the commit; that's expected

**Interfaces:** none — documentation only.

- [ ] **Step 1: Update `AGENTS.md`'s "Dev environment" section.** Replace
  the current text (which still describes mise as running the dev processes
  directly — written before this spec existed) with: per-tool containers
  run the dev servers, built from each tool's own Dockerfile `dev` stage via
  `docker-compose.yaml`; `mise.toml` exists and its job is editor
  tooling on WSL only (type-checking, linting, autocomplete), not running
  anything. Link to
  `docs/superpowers/specs/2026-09-12-local-dev-environment-design.md`, the
  way the frontend spec is already linked elsewhere in this file.
- [ ] **Step 2: Confirm `AGENTS.md`'s repo-layout code block.** Check the
  `mise.toml` line's comment still matches reality once Task 1's Node
  decision is known — adjust if it changed what's actually pinned.
- [ ] **Step 3: Update `NOTES.md`'s "Dev experience" section.** Replace the
  bullets written earlier this session (before this spec existed) with a
  short pointer to the new spec, keeping only what's still true stated
  inline (no Traefik locally; no Kubernetes for local dev) and removing
  anything the per-container model has superseded.
- [ ] **Step 4: Re-read both files start to finish.** Confirm no remaining
  reference to a shared `workspace` devcontainer or to mise running dev
  processes. The frontend's nginx-based `prod` stage is correct and stays
  described as such.
- [ ] **Step 5: Commit.**
  ```bash
  git add AGENTS.md
  git commit -m "docs: Describe the finished per-tool dev container model"
  ```
