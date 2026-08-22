# CI/CD — GitHub Actions → VPS

See also: [docs/cicd-setup.md](./docs/cicd-setup.md) in the monorepo docs folder, or the summary below.

## Workflows

| File | Trigger | Action |
|------|---------|--------|
| `.github/workflows/ci-cd.yml` | push / PR / manual | CI on every run; deploy to VPS only after CI passes on `master` |

Deploy includes API health check and optional chained client deploy via `REPO_DISPATCH_TOKEN`.

## Required GitHub secrets

- `SSH_PRIVATE_KEY` — deploy key (private)
- `SSH_HOST` — e.g. `46.249.101.208`
- `SSH_USER` — e.g. `root`
- `SERVER_DEPLOY_PATH` — e.g. `/root/source/Rahil-gallery-backend-go`
- `REPO_DISPATCH_TOKEN` — (optional) GitHub PAT with `repo` scope to auto-deploy the client after API deploy

## First-time VPS

1. Load base images once: handled automatically by CI (`base-images.tar.gz`); manual fallback is `scripts/bundle-docker-images.sh`  
2. Create `.env` from `.env.example` (set `JWT_ACCESS_SECRET`)  
3. Add deploy public key to `authorized_keys`

Production stack: `make docker-prod-up`

If you run compose manually, first export derived production env:

```bash
eval "$(bash scripts/export-compose-env.sh)"
docker compose -f docker-compose.prod.yml up -d
```

## What CI deploys to the VPS

Each deploy syncs the **full tracked source tree** at the deployed commit (via `git archive` + `rsync`), then loads the API Docker image:

- All Go source: `cmd/`, `internal/`, `migrations/`, `scripts/`, compose files, `Makefile`, `docs/`, etc.
- Pre-built Linux binaries: `bin/seed-prod`, `bin/backfill-crm`
- `.deploy-sha` — commit SHA written during sync (check with `cat .deploy-sha`)

Preserved on the server (never overwritten by CI):

- `.env`
- `data/` (catalog images, customer signatures)

CI does **not** run `git pull` on the server — the checkout may be stale until the next deploy syncs it. After deploy, `cat .deploy-sha` should match the latest GitHub commit.

Runtime API code runs from the **Docker image**, not `go run`. The synced source is for migrations, Make targets, seeders, and debugging.

During deploy, `scripts/deploy-remote.sh` also validates database credentials:

- `DATABASE_URL_DOCKER` is always derived from `.env` `DATABASE_URL` when present.
- `POSTGRES_PASSWORD` is derived from `DATABASE_URL` before `db` starts, so fresh volumes initialize with the same password the API will use.
- After `db` is up, deploy tests the `DATABASE_URL` credentials from the same Docker network path used by migrations and the API.
- If Postgres rejects those credentials, deploy fails before migrations/API restart. It does **not** rewrite database role passwords automatically.

If deploy fails with `Postgres rejected the credentials derived from .env DATABASE_URL`, fix the VPS `.env` so `DATABASE_URL` matches the existing database role password, or rotate the role password as a one-time operational action.

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

Required once on VPS: base images are loaded automatically on the first CI deploy. Use `scripts/bundle-docker-images.sh` only for manual/offline bootstrap.
