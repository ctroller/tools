# Platform Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold `apps/frontend/` — a SvelteKit static SPA (landing page, shared nav, one placeholder tool route) that builds and deploys standalone, with no tool-specific logic yet.

**Architecture:** SvelteKit with SSR disabled (`adapter-static`, fallback SPA mode), TypeScript throughout, Bun as runtime/package manager. Pico.css via CDN link in `app.html` for base styling; Svelte's built-in per-component scoped styles for anything else. Client-side routing only — no server-rendering, no auth, no real tool logic (file-converter's actual UI is a separate, later effort once its backend job API exists).

**Tech Stack:** SvelteKit 2 / Svelte 5, TypeScript, Vite, Vitest + @testing-library/svelte, Bun, Docker (multi-stage: Bun build → nginx static serve).

**Spec:** `docs/superpowers/specs/2026-08-30-platform-frontend-design.md`

## Global Constraints

- TypeScript throughout — no plain `.js` source files under `src/`.
- SSR off (`export const ssr = false` in the root layout); `adapter-static` with `fallback: 'index.html'`.
- Bun is the package manager and script runner for every command in this plan (`bun install`, `bun run <script>`), not `npm`/`pnpm`/`yarn`.
- No runtime API base-URL config — the frontend calls relative paths (`/api/<tool>/...`) only. The only place a backend address is configured is the dev-server proxy (`vite.config.ts`, dev-time only).
- Styling: Pico.css via CDN `<link>` in `src/app.html`, plus Svelte's automatic per-component scoped `<style>` blocks. No Tailwind, no Alpine.js.
- Testing: Vitest + `@testing-library/svelte`, not Bun's built-in test runner (per spec — Bun's Svelte component-testing support is less proven).
- No `lib/stores/` (shared async-job store) and no `lib/api/client.ts` in this plan — both are deferred per spec until a real tool backend exists to inform their shape. Do not add them speculatively.
- No linting/formatting tooling added — deferred per spec.

---

### Task 1: Scaffold the SvelteKit project

**Files:**
- Create: `apps/frontend/package.json`
- Create: `apps/frontend/svelte.config.js`
- Create: `apps/frontend/vite.config.ts`
- Create: `apps/frontend/tsconfig.json`
- Create: `apps/frontend/src/app.html`
- Create: `apps/frontend/src/routes/+layout.ts`
- Create: `apps/frontend/.gitignore`

**Interfaces:**
- Produces: a working SvelteKit build pipeline (`bun run dev`, `bun run build`, `bun run test`) that later tasks add routes/components/tests into. No exported functions/types — this task is pure project scaffolding.

- [ ] **Step 1: Install Bun, if not already available**

Run: `bun --version`

If that fails with "command not found":

Run: `curl -fsSL https://bun.sh/install | bash && source ~/.bashrc`

Then re-run `bun --version` and confirm it prints a version number before continuing.

- [ ] **Step 2: Create the project directory and `package.json`**

Create `apps/frontend/package.json`:

```json
{
	"name": "frontend",
	"private": true,
	"version": "0.0.1",
	"type": "module",
	"scripts": {
		"dev": "vite dev",
		"build": "vite build",
		"preview": "vite preview",
		"test": "vitest run",
		"prepare": "svelte-kit sync"
	},
	"devDependencies": {
		"@sveltejs/adapter-static": "^3.0.0",
		"@sveltejs/kit": "^2.0.0",
		"@sveltejs/vite-plugin-svelte": "^4.0.0",
		"@testing-library/svelte": "^5.0.0",
		"jsdom": "^25.0.0",
		"svelte": "^5.0.0",
		"typescript": "^5.0.0",
		"vite": "^5.0.0",
		"vitest": "^2.0.0"
	}
}
```

- [ ] **Step 3: Create `svelte.config.js`**

```javascript
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: false
		})
	}
};

export default config;
```

- [ ] **Step 4: Create `vite.config.ts`**

No dev proxy yet — that's Task 6. Vitest config included now so later tasks can write tests immediately.

```typescript
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		environment: 'jsdom',
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
```

- [ ] **Step 5: Create `tsconfig.json`**

```json
{
	"extends": "./.svelte-kit/tsconfig.json",
	"compilerOptions": {
		"allowJs": true,
		"checkJs": true,
		"esModuleInterop": true,
		"forceConsistentCasingInFileNames": true,
		"resolveJsonModule": true,
		"skipLibCheck": true,
		"sourceMap": true,
		"strict": true,
		"moduleResolution": "bundler"
	}
}
```

- [ ] **Step 6: Create `src/app.html`**

```html
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<link
			rel="stylesheet"
			href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"
		/>
		%sveltekit.head%
	</head>
	<body data-sveltekit-preload-data="hover">
		<div style="display: contents">%sveltekit.body%</div>
	</body>
</html>
```

- [ ] **Step 7: Create `src/routes/+layout.ts`**

Disables SSR for the whole app — required for `adapter-static`'s fallback SPA mode.

```typescript
export const ssr = false;
```

- [ ] **Step 8: Create `apps/frontend/.gitignore`**

```
node_modules/
build/
.svelte-kit/
```

- [ ] **Step 9: Install dependencies and generate SvelteKit's internal types**

Run (from `apps/frontend/`):

```bash
bun install
bunx svelte-kit sync
```

Expected: completes without error; a `.svelte-kit/` directory now exists.

Check which lockfile Bun created — needed verbatim for Task 7's Dockerfile:

```bash
ls bun.lock bun.lockb 2>/dev/null
```

Note the filename that exists (one of the two). Task 7 assumes `bun.lock`; if `bun.lockb` is what actually appears, adjust Task 7's Dockerfile `COPY` line to match.

- [ ] **Step 10: There's nothing to render yet — verify the build pipeline itself**

Run: `bun run build`

Expected: exits 0, and `apps/frontend/build/` now contains at least `index.html` and a `_app/` directory. (The page will be SvelteKit's default 404-ish shell — no routes exist yet, that's expected here.)

- [ ] **Step 11: Commit**

```bash
git add apps/frontend/package.json apps/frontend/svelte.config.js apps/frontend/vite.config.ts apps/frontend/tsconfig.json apps/frontend/src/app.html apps/frontend/src/routes/+layout.ts apps/frontend/.gitignore apps/frontend/bun.lock
git commit -m "feat(frontend): scaffold SvelteKit project"
```

(If Step 9 found `bun.lockb` instead, add that filename in place of `bun.lock`.)

---

### Task 2: Shared header and root layout

**Files:**
- Create: `apps/frontend/src/lib/components/SiteHeader.svelte`
- Create: `apps/frontend/src/lib/components/SiteHeader.test.ts`
- Create: `apps/frontend/src/routes/+layout.svelte`

**Interfaces:**
- Consumes: nothing from Task 1 beyond the build pipeline.
- Produces: `SiteHeader` (default Svelte component export from `$lib/components/SiteHeader.svelte`) — a `<header>` with a link to `/` and the text "Tools". Later tasks (3, 4) render inside `+layout.svelte`'s `{@render children()}` slot, so they don't need to re-render the header themselves.

- [ ] **Step 1: Write the failing test**

Create `apps/frontend/src/lib/components/SiteHeader.test.ts`:

```typescript
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import SiteHeader from './SiteHeader.svelte';

describe('SiteHeader', () => {
	it('renders a link back to the tool list', () => {
		render(SiteHeader);
		const link = screen.getByRole('link', { name: 'Tools' });
		expect(link).toBeTruthy();
		expect(link.getAttribute('href')).toBe('/');
	});
});
```

- [ ] **Step 2: Run the test and verify it fails**

Run (from `apps/frontend/`): `bun run test`

Expected: FAIL — `SiteHeader.svelte` does not exist yet (module resolution error).

- [ ] **Step 3: Implement `SiteHeader.svelte`**

Create `apps/frontend/src/lib/components/SiteHeader.svelte`:

```svelte
<header>
	<nav>
		<ul>
			<li><strong><a href="/">Tools</a></strong></li>
		</ul>
	</nav>
</header>
```

- [ ] **Step 4: Run the test and verify it passes**

Run: `bun run test`

Expected: PASS.

- [ ] **Step 5: Wire it into the root layout**

Create `apps/frontend/src/routes/+layout.svelte`:

```svelte
<script lang="ts">
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	let { children } = $props();
</script>

<SiteHeader />
<main>
	{@render children()}
</main>
```

- [ ] **Step 6: Manually verify the layout renders**

Run: `bun run dev`

Open `http://localhost:5173` in a browser. Expected: a "Tools" header link is visible above an empty main area (no page content yet — that's Task 3). Stop the dev server (Ctrl+C) once confirmed.

**If you have no browser tool available** (true for most implementer subagents): say so plainly in your report rather than fabricating this check. The `SiteHeader` component test already proves the component itself renders correctly; this step exists to catch a layout-wiring mistake a unit test can't see (e.g. the component silently not being included). Substitute what you can verify — confirm the dev server starts without error — and note the visual check as not performed in this environment.

- [ ] **Step 7: Commit**

```bash
git add apps/frontend/src/lib/components/SiteHeader.svelte apps/frontend/src/lib/components/SiteHeader.test.ts apps/frontend/src/routes/+layout.svelte
git commit -m "feat(frontend): add shared site header and root layout"
```

---

### Task 3: Landing page

**Files:**
- Create: `apps/frontend/src/routes/+page.svelte`
- Create: `apps/frontend/src/routes/page.test.ts`

**Interfaces:**
- Consumes: nothing new — this is a leaf route rendered inside Task 2's `+layout.svelte`.
- Produces: nothing consumed by later tasks (Task 4 is an independent route). Establishes the link target `/tools/file-converter` that Task 4's route must exist at.

- [ ] **Step 1: Write the failing test**

Create `apps/frontend/src/routes/page.test.ts`:

```typescript
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Page from './+page.svelte';

describe('landing page', () => {
	it('links to the file-converter tool', () => {
		render(Page);
		const link = screen.getByRole('link', { name: 'File Converter' });
		expect(link).toBeTruthy();
		expect(link.getAttribute('href')).toBe('/tools/file-converter');
	});
});
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `bun run test`

Expected: FAIL — `+page.svelte` does not exist yet.

- [ ] **Step 3: Implement the landing page**

Create `apps/frontend/src/routes/+page.svelte`:

```svelte
<h1>Tools</h1>
<ul>
	<li><a href="/tools/file-converter">File Converter</a></li>
</ul>
```

- [ ] **Step 4: Run the test and verify it passes**

Run: `bun run test`

Expected: PASS (both this test and Task 2's `SiteHeader` test).

- [ ] **Step 5: Commit**

```bash
git add apps/frontend/src/routes/+page.svelte apps/frontend/src/routes/page.test.ts
git commit -m "feat(frontend): add landing page"
```

---

### Task 4: File-converter placeholder route

**Files:**
- Create: `apps/frontend/src/routes/tools/file-converter/+page.svelte`
- Create: `apps/frontend/src/routes/tools/file-converter/page.test.ts`

**Interfaces:**
- Consumes: nothing — independent leaf route.
- Produces: nothing consumed by later tasks. This is intentionally a placeholder — no upload UI, no API calls. Real file-converter UI is out of scope for this plan (per spec: built alongside that tool's own backend, once its job API exists).

- [ ] **Step 1: Write the failing test**

Create `apps/frontend/src/routes/tools/file-converter/page.test.ts`:

```typescript
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Page from './+page.svelte';

describe('file-converter placeholder page', () => {
	it('renders a heading and a not-yet-built notice', () => {
		render(Page);
		expect(screen.getByRole('heading', { name: 'File Converter' })).toBeTruthy();
		expect(screen.getByText('Coming soon.')).toBeTruthy();
	});
});
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `bun run test`

Expected: FAIL — route file does not exist yet.

- [ ] **Step 3: Implement the placeholder page**

Create `apps/frontend/src/routes/tools/file-converter/+page.svelte`:

```svelte
<h1>File Converter</h1>
<p aria-busy="true">Coming soon.</p>
```

- [ ] **Step 4: Run the test and verify it passes**

Run: `bun run test`

Expected: PASS — all tests from Tasks 2, 3, and 4 green.

- [ ] **Step 5: Manually verify client-side navigation**

Run: `bun run dev`

Open `http://localhost:5173`, click "File Converter". Expected: URL becomes `/tools/file-converter`, page shows "File Converter" heading and "Coming soon.", header stays visible (confirms the shared layout persists across client-side navigation). Stop the dev server.

**If you have no browser tool available:** say so plainly rather than fabricating this check — do not invent evidence. The landing-page test (Task 3) already proves the link's `href` is correct, and the placeholder-page test (this task) already proves that route's content renders; what a unit test genuinely can't prove is that SvelteKit's client-side router and shared layout actually persist across navigation in a real browser. That residual risk is accepted as covered by SvelteKit's standard, framework-guaranteed routing behavior (not something that could work for other routes and fail for this one) rather than by an empirical check no available tool can perform.

- [ ] **Step 6: Commit**

```bash
git add apps/frontend/src/routes/tools/file-converter/+page.svelte apps/frontend/src/routes/tools/file-converter/page.test.ts
git commit -m "feat(frontend): add file-converter placeholder route"
```

---

### Task 5: Error page for unhandled route/load failures

**Files:**
- Create: `apps/frontend/src/routes/+error.svelte`
- Create: `apps/frontend/src/routes/error.test.ts`

**Interfaces:**
- Consumes: `page` store from `$app/stores` (SvelteKit built-in — provides `$page.status: number` and `$page.error: { message: string } | null` on an unhandled route/load error).
- Produces: nothing consumed by later tasks.

- [ ] **Step 1: Write the failing test**

Create `apps/frontend/src/routes/error.test.ts`:

```typescript
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import { readable } from 'svelte/store';

vi.mock('$app/stores', () => ({
	page: readable({ status: 404, error: { message: 'Not Found' } })
}));

import ErrorPage from './+error.svelte';

describe('+error.svelte', () => {
	it('shows the status code and error message', () => {
		render(ErrorPage);
		expect(screen.getByText('404')).toBeTruthy();
		expect(screen.getByText('Not Found')).toBeTruthy();
	});
});
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `bun run test`

Expected: FAIL — `+error.svelte` does not exist yet.

- [ ] **Step 3: Implement the error page**

Create `apps/frontend/src/routes/+error.svelte`:

```svelte
<script lang="ts">
	import { page } from '$app/stores';
</script>

<h1>{$page.status}</h1>
<p>{$page.error?.message}</p>
```

- [ ] **Step 4: Run the test and verify it passes**

Run: `bun run test`

Expected: PASS — all tests from Tasks 2–5 green.

- [ ] **Step 5: Manually verify it fires for a real unmatched route**

Run: `bun run dev`

Open `http://localhost:5173/does-not-exist`. Expected: the page shows `404` and a not-found message, styled by Pico (not a raw browser error page). Stop the dev server.

**If you have no browser tool available:** say so plainly rather than fabricating this check — do not invent evidence. It's genuinely unclear without testing whether SvelteKit's dev server SSRs this route despite `ssr = false` at the root layout (dev and build don't necessarily behave identically here); if a `curl` against the dev server happens to show real rendered content, report that as a real (if unplanned) data point rather than assuming it will. Either way, the `+error.svelte` component test already proves the component itself renders the right status/message given mocked store data — that's the part a unit test can cover regardless of what curl shows.

- [ ] **Step 6: Commit**

```bash
git add apps/frontend/src/routes/+error.svelte apps/frontend/src/routes/error.test.ts
git commit -m "feat(frontend): add error page for unhandled route failures"
```

---

### Task 6: Dev-server API proxy

**Files:**
- Modify: `apps/frontend/vite.config.ts`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing consumed by later tasks — this is dev-tooling config only, verified against a throwaway dummy server, not against any real gateway (the gateway doesn't exist yet in this repo).

- [ ] **Step 1: Add the proxy config**

Modify `apps/frontend/vite.config.ts` — add a `server.proxy` block:

```typescript
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		environment: 'jsdom',
		include: ['src/**/*.{test,spec}.{js,ts}']
	},
	server: {
		proxy: {
			'/api': {
				target: process.env.API_PROXY_TARGET ?? 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
```

- [ ] **Step 2: Start a throwaway target server to prove the proxy forwards correctly**

In a separate terminal, from any directory:

```bash
mkdir -p /tmp/dummy-api && cd /tmp/dummy-api
echo '{"ok":true}' > ping.json
bunx --bun http-server -p 8080 --cors
```

(If `http-server` isn't available, `python3 -m http.server 8080` from `/tmp/dummy-api` works equally well as the dummy target — either serves `ping.json` as a static file.)

- [ ] **Step 3: Start the frontend dev server and verify the proxy**

From `apps/frontend/`, in another terminal:

```bash
bun run dev &
sleep 1
curl -s http://localhost:5173/api/ping.json
```

Expected output: `{"ok":true}` — proves a request to the frontend dev server under `/api/*` was forwarded to the dummy target on port 8080, not handled by SvelteKit itself.

Stop both the dev server and the dummy target server (`kill %1` or Ctrl+C each).

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/vite.config.ts
git commit -m "feat(frontend): add dev-server proxy for /api"
```

---

### Task 7: Dockerfile — multi-stage build and static serve

**Files:**
- Create: `apps/frontend/Dockerfile`
- Create: `apps/frontend/nginx.conf`
- Create: `apps/frontend/.dockerignore`

**Interfaces:**
- Consumes: the build output of `bun run build` (from Task 1's `package.json` `build` script) and whichever lockfile Task 1 produced.
- Produces: a runnable container image; nothing else in this plan depends on it.

- [ ] **Step 1: Create `apps/frontend/.dockerignore`**

```
node_modules
build
.svelte-kit
```

- [ ] **Step 2: Create `apps/frontend/nginx.conf`**

Serves the static build output and falls back to `index.html` for any path SvelteKit's client-side router owns (e.g. `/tools/file-converter` on a hard refresh) — the same fallback the `adapter-static` config in Task 1 produces.

```nginx
server {
	listen 80;
	root /usr/share/nginx/html;
	index index.html;

	location / {
		try_files $uri $uri/ /index.html;
	}
}
```

- [ ] **Step 3: Create `apps/frontend/Dockerfile`**

Uses `bun.lock` per Task 1's Step 9 check — if that found `bun.lockb` instead, change the `COPY` line below to match before building.

```dockerfile
FROM oven/bun:1 AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

- [ ] **Step 4: Build the image**

Run (from `apps/frontend/`):

```bash
docker build -t frontend-test .
```

Expected: exits 0.

- [ ] **Step 5: Run it and verify both a static route and a client-only route are served**

**Correction (found during execution, see the plan's execution ledger):** the original version of this step expected `curl | grep -o 'Tools'` to succeed. That's impossible as written — with `ssr = false` (locked in Task 1), the static build's `index.html` is an unhydrated shell with no server-rendered text at all; "Tools" only appears after client-side JS executes in a real browser, which `curl` can never do. The corrected check below verifies the thing this step actually needs to prove — that nginx's `try_files` fallback serves the *same* shell for an unmapped client-only route instead of a raw 404 — without relying on rendered text.

```bash
docker run -d --rm -p 8081:80 --name frontend-test frontend-test
sleep 1
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8081/
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8081/tools/file-converter
curl -s http://localhost:8081/ | md5sum
curl -s http://localhost:8081/tools/file-converter | md5sum
docker stop frontend-test
```

Expected: both status codes are `200`, and both `md5sum` outputs match each other exactly. That match is the proof — it means nginx served the identical `index.html` file for both the real landing-page path and the client-only route (via the `try_files` fallback), rather than a distinct real file for one and a 404 for the other.

- [ ] **Step 6: Commit**

```bash
git add apps/frontend/Dockerfile apps/frontend/nginx.conf apps/frontend/.dockerignore
git commit -m "feat(frontend): add multi-stage Dockerfile"
```

---

## Explicitly not in this plan

Matches the spec's "Deferred" and "Out of scope" sections — do not add these speculatively:

- `lib/stores/` (shared async-job store) — no backend job API exists yet to shape it.
- `lib/api/client.ts` — no consumer exists yet; nothing in this plan makes a real API call.
- Any actual file-converter UI (upload, format selection, status/download) — separate future work alongside that tool's backend.
- Auth.
- Linting/formatting tooling.
- Devcontainer/`compose-workspace.yaml` wiring for Bun — those files don't exist yet in this repo at all (pre-existing gap, not created by this plan); running this plan's commands assumes Bun is available in whatever shell executes it.
