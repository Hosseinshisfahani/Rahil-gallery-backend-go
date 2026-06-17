#!/usr/bin/env bash
# Server-side deploy hook — called by GitHub Actions over SSH.
# Usage: bash scripts/deploy-remote.sh /tmp/api-image.tar.gz <git-sha>
set -euo pipefail

IMAGE_TAR="${1:?image tar path required}"
GIT_SHA="${2:-latest}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

API_IMAGE="${API_IMAGE:-rahil-gallery-api:${GIT_SHA}}"

echo "==> Loading API image from ${IMAGE_TAR}"
gunzip -c "${IMAGE_TAR}" | docker load

LOADED="$(docker images --format '{{.Repository}}:{{.Tag}}' | grep '^rahil-gallery-api:' | head -1 || true)"
if [[ -z "${LOADED}" ]]; then
  echo "ERROR: rahil-gallery-api image not found after docker load" >&2
  exit 1
fi

docker tag "${LOADED}" rahil-gallery-api:latest
export API_IMAGE="rahil-gallery-api:latest"

if [[ ! -f .env ]]; then
  echo "WARN: .env missing — copy from .env.example and set JWT_ACCESS_SECRET" >&2
fi

echo "==> Ensuring base images (postgres, migrate) exist locally"
if ! docker image inspect "${POSTGRES_IMAGE:-postgres:16-alpine}" >/dev/null 2>&1; then
  echo "Postgres image missing. Run once: bash scripts/bundle-docker-images.sh (on a machine with Docker Hub), then load on this server."
  exit 1
fi
if ! docker image inspect "${MIGRATE_IMAGE:-migrate/migrate:v4.18.1}" >/dev/null 2>&1; then
  echo "Migrate image missing. Run once: bash scripts/bundle-docker-images.sh (on a machine with Docker Hub), then load on this server."
  exit 1
fi

echo "==> Ensuring database is up"
docker compose -f docker-compose.prod.yml up -d db

echo "==> Running migrations"
docker compose -f docker-compose.prod.yml up migrate --abort-on-container-exit

echo "==> Starting API"
docker compose -f docker-compose.prod.yml up -d api

echo "==> Pruning old API images (keep latest + current sha)"
docker images rahil-gallery-api --format '{{.ID}} {{.Tag}}' \
  | awk -v keep="${GIT_SHA}" '$2 != "latest" && $2 != keep {print $1}' \
  | xargs -r docker rmi -f 2>/dev/null || true

rm -f "${IMAGE_TAR}"
echo "==> Deploy complete (${GIT_SHA})"
