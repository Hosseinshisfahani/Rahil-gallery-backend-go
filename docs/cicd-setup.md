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
