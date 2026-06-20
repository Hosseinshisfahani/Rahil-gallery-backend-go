# CI/CD — GitHub Actions → VPS

See also: [docs/cicd-setup.md](./docs/cicd-setup.md) in the monorepo docs folder, or the summary below.

## Workflows

| File | Trigger | Action |
|------|---------|--------|
| `.github/workflows/ci.yml` | push / PR | Unit tests + Docker build |
| `.github/workflows/deploy.yml` | push to `master` | Build image → SCP to VPS → `scripts/deploy-remote.sh` |

## Required GitHub secrets

- `SSH_PRIVATE_KEY` — deploy key (private)
- `SSH_HOST` — e.g. `46.249.101.208`
- `SSH_USER` — e.g. `root`
- `SERVER_DEPLOY_PATH` — e.g. `/root/source/Rahil-Gallery-Server`

## First-time VPS

1. Load base images once: `bash scripts/bundle-docker-images.sh` locally → `docker load` on server  
2. Create `.env` from `.env.example` (set `JWT_ACCESS_SECRET`)  
3. Add deploy public key to `authorized_keys`

Production stack: `docker compose -f docker-compose.prod.yml up -d`

## What CI deploys to the VPS

CI does **not** run `git pull` on the server. Each deploy uploads:

- Docker API image (built in GitHub Actions)
- `migrations/`, compose files, deploy scripts, `Makefile`
- Pre-built Linux binaries: `bin/seed-prod`, `bin/backfill-crm`

The git checkout under `SERVER_DEPLOY_PATH` may be outdated — that is normal. Use `make seed-prod` / `make seed-backfill-crm` (they use `bin/*` when present), not `go run ./cmd/...`.

To refresh the full repo for local development on the VPS:

```bash
cd /root/source/Rahil-Gallery-Server
git stash push -u -m deploy-artifacts   # keep .env
git pull origin master
git stash pop || true
```

## VPS without Docker Hub

`make docker-up` **builds on the server** and needs `golang:1.22-alpine` from Docker Hub — it will fail with TLS timeout on blocked networks.

Use one of these instead:

| Goal | Command |
|------|---------|
| DB + migrations only, API via Go | `make docker-dev` then see below |
| Go on VPS (try mirror first) | `export GOPROXY=https://goproxy.io,direct && make run` |
| Go on VPS (offline) | Laptop: `bash scripts/bundle-api-binary.sh` → SCP → `./bin/rahil-api` |
| Full stack with containerized API | CI/CD deploy, then `make docker-prod-up` |
| Manual API image transfer | build locally → `docker save` → `docker load` on VPS → `make docker-prod-up` |

Required once on VPS: load base images via `scripts/bundle-docker-images.sh`.
