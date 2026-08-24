# AGENT.md

This file provides guidance agents when working with code in this repository.

## What this is

Personal homelab toolbox platform. The primary goal is improving cloud-native skills (k3s, ingress, GitOps); the tools themselves are genuinely useful but the platform is intentionally disposable. NOTES.md is the authoritative re-entry document — read it when picking up work after a gap.

## Repo layout

```
apps/<tool>/          # one service per tool (own Dockerfile)
deploy/<tool>/        # k8s manifests for that tool (mirrors apps/ layout)
compose-workspace.yaml  # dev environment (workspace devcontainer + tool services)
.devcontainer/        # VS Code devcontainer config
```

`deploy/` is structured for GitOps (future Argo/Flux points here, no reshuffle needed).

## Tools

Each tool is a standalone app under `apps/<tool>/`. See individual README/AGENTS files for more info.

### Watch-outs for services
- Services must bind `0.0.0.0`, not `127.0.0.1` — localhost bind = unreachable = 502 that looks like a routing bug.

## Dev environment

Start the workspace with Docker Compose:

```bash
docker compose -f compose-workspace.yaml up -d
```

Attach VS Code to the `workspace` service (devcontainer). Tool services (e.g. `file-converter`) are defined as separate Compose services alongside `workspace` — uncomment them in `compose-workspace.yaml` as tools are built.

Devcontainer ships: Go 1.26, Node 26, kubectl. No Helm or Minikube.

## Routing contract (prod and dev parity)

Public base: `tools.trox.dev`
- Frontend: `/`
- Each tool backend: `/api/<tool>/`

**StripPrefix** Traefik middleware strips `/api/<tool>` before forwarding, so each service receives `/` and stays ignorant of its public path. Dev Compose must mirror this contract — frontend reads `API_BASE_URL` from env, never hardcode ports.

## Kubernetes / deploy
Manifests in `deploy/<tool>/` are split by kind (Deployment, Service, Ingress, Middleware) to stay GitOps-ready. No Kustomize overlays until there's a second deploy target.

## History

Keep a history under .claude/

## Agent behaviour

- Unless otherwise instructed, you are not to give the user code examples or a finished solution. You are primarily an architect, code reviewer, or tester. The only exclusion to this rule is Frontend code, but even there you should provide pair programming advice, not full solutions.
- **Always use ASD-STE100 Simplified Technical English**
- When writing something intended for human consumption, (comment, commit message, reply to prompt) use as few words as possible. Pick every word meticulously to reduce the volume to a strict minimum. Be down to the point. Less is more.
- Avoid superlatives and praise. Stop telling me I am absolutely right. Give me the cold hard truth.
- Let the reader of the code breathe. Add empty lines between logical blocks of code.