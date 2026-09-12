# Local dev environment — design

## Status

Design agreed, not yet implemented. `compose-workspace.yaml` exists but is
empty. `.mise.toml` does not exist yet. No tool's Dockerfile has a `dev` stage
yet.

## Context

`AGENTS.md` and `NOTES.md` originally locked "Option A: workspace devcontainer
+ Docker Compose" — one devcontainer holding the editing toolchains, attached
to by VS Code, with tool services as Compose siblings. That plan was reopened
once the actual editor setup came up: VS Code runs on Windows, attaching to a
repo checkout inside WSL2. WSL2 already is a real Linux kernel, so a
devcontainer around the toolchain added an editor-attach hop on top of a
problem WSL2 had already solved — no remaining reason to keep it.

That produced an interim decision: install the toolchain directly into WSL,
version-pinned per-project with **mise**, Docker Compose kept only for
services that must run as containers. Revisiting the frontend's Dockerfile
against that decision then surfaced the actual target: the difference between
dev and prod should be as small as possible, live reload and debugger attach
need to work, and the answer has to generalize as more (possibly polyglot)
tools get added — not just fit the frontend.

This spec replaces the "Dev experience" section of `NOTES.md` and the earlier
mise-runs-everything plan. mise's role survives, narrowed to editor tooling
only.

This spec also supersedes one line in
`2026-08-30-platform-frontend-design.md`'s "Build / package manager / deploy"
section: the frontend's `prod` stage serves its own build output via Bun,
not nginx/Caddy — nginx was dropped as an unnecessary dependency once every
other decision here settled on per-tool containers built from a single
pinned base image.

## Decisions

### Split of responsibilities: running process vs. editor tooling

Two different consumers of a tool's toolchain, two different homes:
- **Running process** (dev server, build, hot reload) — a per-tool container,
  built from that tool's own Dockerfile.
- **Editor tooling** (type checking, autocomplete, go-to-definition) — mise on
  WSL directly, pinned to match each Dockerfile's version exactly.

Rejected:
- **A single shared devcontainer for all toolchains** — duplicates WSL2's own
  Linux capability, adds an editor-attach hop with no remaining problem left
  to solve, and its main benefit (identical toolchain for any future
  contributor) has little value for a single-user project.
- **Uniform Compose with no host-side toolchain at all** — breaks editor
  tooling, since WSL's language servers need the toolchain and dependencies
  visible on the WSL filesystem to resolve against.

### Per-tool Dockerfile shape

Each `apps/<tool>/Dockerfile` carries a `dev` stage next to its `prod` stage,
both built from the same pinned base image tag — one version, two jobs.
`docker build --target dev` / `--target prod` (or Compose's `build.target`)
selects which.

- **Go tools**: `dev` stage adds a hot-rebuild tool (e.g. `air`), run against
  bind-mounted source. `prod` stage keeps the existing static-binary build —
  `file-converter`'s current Dockerfile is the working example of that half.
- **Frontend**: `dev` stage runs `bun run dev --host 0.0.0.0` (binding
  matters — see `AGENTS.md`'s `0.0.0.0` watch-out, which exists for exactly
  this class of bug). `prod` stage builds via Bun and serves the static
  output via Bun directly; nginx is dropped entirely.

### Live reload and debugging

The repo lives on the WSL filesystem, so a bind mount from WSL into a Compose
container stays Linux-to-Linux — no cross-filesystem `inotify` gap. Hot-reload
tooling (`air`, Vite's HMR) should see real file-change events, not need a
polling fallback. The usual complaint about containerized dev being slow to
pick up file changes is a Windows-drvfs problem; it does not apply here.

Debugger attach is a forwarded port, nothing more:
- Go: Delve, `dlv --headless --listen=:2345`.
- Frontend: Bun's `--inspect`.

The editor attaches to `localhost:<forwarded-port>` and never enters the
container — no Dev Containers extension, no remote-attach-into-container
flow.

### Dev networking

No Traefik locally — unchanged from `NOTES.md`'s routing decision, prod-only.
Containers reach each other by Compose service name over the Compose network
(e.g. `file-converter`), not `localhost`.

`vite.config.ts` already supports this, unused until now: `API_PROXY_TARGET`
overrides the dev-server proxy target (default `http://localhost:8080`).
Compose sets `API_PROXY_TARGET=http://file-converter:8080` for the frontend's
dev service — no frontend code change needed for this.

### Orchestration

`compose-workspace.yaml` becomes the one place every tool's dev container is
defined — one service per tool, each built with `target: dev`. `docker
compose -f compose-workspace.yaml up -d` starts the world. The
`Makefile`/`justfile` wrapper (`up`/`down`/`logs`) named in `NOTES.md` still
applies, now fronting per-tool containers instead of a shared workspace
container.

### Editor tooling via mise

`.mise.toml` at the repo root pins Go, Node, Bun, and kubectl for WSL,
versions kept in sync with whatever each tool's Dockerfile pins. This is a
deliberate, small duplication — a version bump touches two files, the
Dockerfile and `.mise.toml` — accepted in exchange for full editor support
without an editor-attach hop into any container. kubectl has no per-tool
container equivalent; it stays a host-level CLI, pinned the same way.

## Deferred (explicitly, not forgotten)

- Named volumes over `node_modules` / `.svelte-kit` (and Go's module cache)
  inside each tool's bind mount, so a container-side install doesn't collide
  with whatever mise installed on the WSL side. Needed once a dev container
  is actually built, not before.
- Exact hot-rebuild tool for Go (`air` named as the working assumption, not
  committed — `reflex` or another option stays open).
- Non-root user for the frontend's `dev` stage. The `prod` stage's non-root
  decision (from the earlier Dockerfile review) still applies; whether `dev`
  needs the same hardening, given it only ever runs on a trusted local
  machine, is not decided.
- The gateway's dev/prod container shape — the gateway itself is still
  deferred per `NOTES.md`; this spec's per-tool contract applies to it once
  it exists, not before.

## Out of scope

- The gateway's own design (routing logic, auth) — a separate spec once it
  gets built.
- Any individual tool's application code or UI — this spec covers only how
  each tool's dev/prod container pair is shaped, not what runs inside it.
- Kubernetes / `deploy/<tool>/` manifests — unchanged by this spec; this
  covers `apps/<tool>/` and the dev loop only.
