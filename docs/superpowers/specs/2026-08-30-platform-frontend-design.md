# Platform frontend — design

## Status

Design agreed, not yet implemented. `apps/frontend/` does not exist yet.

## Context

The platform serves one frontend for all tools, at `/`, alongside each
tool's own backend at `/api/<tool>/` (routing contract in the root
`AGENTS.md`). This spec covers that frontend only — tool backends are
each their own spec (see `apps/file-converter/ARCHITECTURE.md`).

Starting point was a deferred discussion in
`.claude/history/frontend-notes.md` (flat multi-page HTML, Alpine.js, no
build step). That direction changed once two new constraints came up:
TypeScript is wanted, and the tool count may grow to 10+. TypeScript
forces a build step regardless, which removed the main reason the
no-build-step plan was attractive — so the whole approach was
re-decided from there, not patched.

## Decisions

### Framework: SvelteKit

Considered: keeping Alpine.js + adding TS around it, Lit (Web
Components, TS-native), Vue, Svelte, SolidJS, Angular, a full SPA in
general vs. staying multi-page HTML.

Rejected specifically:
- **Alpine.js + TS** — most of Alpine's logic wiring lives in HTML
  attribute strings (`x-data="..."`), which the TS compiler can't see
  into. TypeScript's value (checked bindings, safe refactors) would be
  mostly lost.
- **React** — explicitly ruled out by preference (steer clear of the
  dominant/most-boilerplate option; the other frameworks work similarly
  enough in practice that the differentiation isn't worth it here).
- **Angular** — full batteries-included (DI, forms, routing, HTTP
  client, opinionated structure). Its value is enforcing structure on
  large teams/apps; oversized for a handful of tool pages.
- **Vue / SolidJS** — reasonable alternatives, not chosen. SolidJS
  (signals-based reactivity) was the runner-up, worth reconsidering
  later if the signals model becomes specifically interesting — Svelte
  5 has since moved toward the same idea via runes.
- **Multi-page static HTML** — the original plan. Reopened once a
  build step became unavoidable (TS) and scale (10+ tools) made a real
  component/routing model worth having.
- **Lit** — considered earlier as "keep custom elements, add TS," before
  the decision to go full SPA. Superseded once a real app framework
  (routing, TS-native components) was on the table anyway — Lit itself
  has no built-in router, so it would've needed one bolted on for the
  same routing need SvelteKit gives natively.

Chosen: **Svelte**, via **SvelteKit** (file-based routing, TS wired in,
builds on Vite underneath) rather than plain Svelte + a router library
— SvelteKit is what current Svelte docs/tutorials target, so it's the
better-supported path.

**SSR is off** (`adapter-static`, static SPA output). No SEO or
server-rendering need — this is a tool dashboard behind unguessable
handles, not public content.

### Directory structure

```
apps/frontend/
  src/
    routes/
      +layout.svelte          # shared shell: <SiteHeader>, Pico link, global nav
      +page.svelte             # landing page — tool list
      tools/
        file-converter/
          +page.svelte          # file-converter UI
    lib/
      components/
        SiteHeader.svelte
      api/
        client.ts                # typed fetch wrapper
    app.html                   # root HTML shell (Pico CDN link here)
  static/                     # favicon, static assets
  svelte.config.js            # adapter-static
  vite.config.ts              # dev proxy: /api/* -> gateway
  package.json
  Dockerfile                  # multi-stage: Bun build -> static serve
```

`lib/stores/` for shared cross-tool logic (e.g. an upload -> poll ->
status -> download state machine) is anticipated but **not created
yet** — see "Deferred" below.

### Routing

One route per tool under `/tools/<name>` (e.g. `/tools/file-converter`),
close enough to the backend's `/api/<tool>/` shape to stay intuitive
without needing to match exactly. Landing page (`/`) lists tools.

### Styling

**Pico.css** (classless CDN framework) as the base/reset layer —
typography, buttons, forms, and native `aria-busy="true"` support for
loading states, which maps directly onto the job-status flow the tool
backends will expose. Anything tool-specific on top uses Svelte's
built-in per-component scoped styles (each `.svelte` file's `<style>`
block is automatically scoped by the compiler) — this gets the style
isolation that was originally considered via Shadow Dom + Web
Components, without the Alpine/Shadow-DOM friction that approach would
have had (moot now that Alpine is dropped, but the isolation property
still holds under Svelte for free).

### API access

Prod serves frontend and every tool backend from the same origin
(`tools.trox.dev/`, `/api/<tool>/` via Traefik), so the frontend uses
**relative fetch paths** everywhere (`/api/file-converter/...`) — no
runtime base-URL configuration needed at all. This replaces the
`API_BASE_URL` env-var plan from the root `NOTES.md`.

The only place a backend URL needs to be known is `vite.config.ts`'s
dev-server proxy target (`server.proxy`), which is a build/dev-time
setting, never shipped to the browser. This gives dev the same
same-origin behavior as prod without CORS handling.

`lib/api/client.ts` holds no base-URL logic, then — its job is a thin
typed wrapper around `fetch` for relative paths (JSON parsing, shared
error handling for the backend's `Problem+json` shape). Whether it's
worth having at all versus calling `fetch` directly per tool is small
enough to decide during implementation, not here.

### Build / package manager / deploy

**Bun** — runtime, package manager, and bundler-adjacent tooling.
Chosen over pnpm (the safer, more proven alternative) by explicit
preference, despite pnpm being flagged as lower-risk: Bun was acquired
by Anthropic in December 2025; stays MIT/open-source with stated
continued investment, but post-acquisition roadmap direction (general
web-dev ecosystem vs. AI-coding-workflow-specific tooling) is still
unproven this soon after the deal. Noted as an accepted risk, not an
oversight.

Dockerfile is multi-stage:
1. Bun stage: install deps, `bun run build` (SvelteKit static output)
2. Runtime stage: nginx/caddy, copy static output only

Mirrors the `apps/<tool>/` — own-Dockerfile convention from the root
`NOTES.md`, even though the frontend isn't a "tool" in the registry
sense.

### Testing

**Vitest**, not Bun's built-in test runner, despite Bun being the
runtime/package manager elsewhere. SvelteKit's component-testing
tooling (`@testing-library/svelte`, jsdom setup) is documented and
maintained against Vitest; Bun's own test runner's Svelte support is
less proven. Revisit once that changes — not a permanent decision.

Scope: component/unit tests only for now. End-to-end (Playwright) is a
named idea, not decided — mirrors the tool backend's own "testing
should exist" non-decision in `apps/file-converter/ARCHITECTURE.md`.

### Error handling

`+error.svelte` per route segment for unhandled failures (route/load
errors). API-level error handling (consuming the backend's
`Problem+json` shape) is each tool's own concern, built alongside that
tool's UI — not a platform-frontend-wide mechanism yet.

## Deferred (explicitly, not forgotten)

- **Shared async-job store** (`lib/stores/` — upload -> poll/SSE ->
  status -> download state machine reusable across tools). This is the
  answer to repeated-logic-per-tool at 10+ tools, but the backend job
  API it would wrap isn't built yet (`apps/file-converter/ARCHITECTURE.md`
  still has the HTTP layer unbuilt). Build this once the first tool's
  backend job API actually exists, informed by its real shape rather
  than guessed.
- **Playwright / end-to-end testing** — named, not scheduled.
- **Bun as the test runner** — revisit once its Svelte/SvelteKit
  component-testing support matures.
- Exact route-naming (`/tools/<name>` vs. flatter `/<name>`) is a
  low-stakes detail, not load-bearing — change freely if it stops
  feeling right once more tools exist.
- **Linting/formatting** — not decided; use whatever the SvelteKit
  scaffold tool defaults to, revisit only if that default is actually
  unpleasant to work with.

## Out of scope

- Any individual tool's UI (file-converter's screens, etc.) — each
  tool's frontend gets designed alongside its own backend.
- Auth / accounts — not part of the platform yet (tool backends use
  unguessable handles, no login).
