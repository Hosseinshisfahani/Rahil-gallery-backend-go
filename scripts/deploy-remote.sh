#!/usr/bin/env bash
# Server-side deploy hook — called by GitHub Actions over SSH.
# Usage: bash scripts/deploy-remote.sh /tmp/api-image.tar.gz <git-sha>
set -euo pipefail

IMAGE_TAR="${1:?image tar path required}"
GIT_SHA="${2:-latest}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

API_IMAGE="${API_IMAGE:-rahil-gallery-api:${GIT_SHA}}"
OBSERVABILITY_ENABLED="${OBSERVABILITY_ENABLED:-true}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:16-alpine}"
MIGRATE_IMAGE="${MIGRATE_IMAGE:-migrate/migrate:v4.18.1}"
PROMETHEUS_IMAGE="${PROMETHEUS_IMAGE:-prom/prometheus:v2.55.1}"
GRAFANA_IMAGE="${GRAFANA_IMAGE:-grafana/grafana:11.4.0}"
COMPOSE=(docker compose -f docker-compose.prod.yml)
COMPOSE_OBS=(docker compose -f docker-compose.prod.yml -f docker-compose.observability.yml)

require_image() {
  local image="$1"
  local hint="$2"
  if ! docker image inspect "${image}" >/dev/null 2>&1; then
    echo "ERROR: ${image} not found locally." >&2
    echo "       ${hint}" >&2
    exit 1
  fi
}

load_env() {
  if [[ ! -f .env ]]; then
    return
  fi
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
}

# Migrate/API use host `db` inside Compose. Always derive from DATABASE_URL when set
# so a stale DATABASE_URL_DOCKER in .env cannot break deploy.
prepare_database_url_docker() {
  if [[ -n "${DATABASE_URL:-}" ]]; then
    export DATABASE_URL_DOCKER="$(
      printf '%s' "$DATABASE_URL" | sed -E 's/(postgres(ql)?:\/\/[^@]+@)[^:/]+/\1db/'
    )"
    echo "==> Using DATABASE_URL_DOCKER derived from DATABASE_URL"
    return
  fi
  if [[ -n "${DATABASE_URL_DOCKER:-}" ]]; then
    export DATABASE_URL_DOCKER
    echo "==> Using DATABASE_URL_DOCKER from .env"
    return
  fi
  echo "ERROR: set DATABASE_URL (recommended) or DATABASE_URL_DOCKER in .env" >&2
  exit 1
}

load_env
prepare_database_url_docker

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
  echo "WARN: .env missing — copy from .env.example and set JWT_ACCESS_SECRET + DATABASE_URL" >&2
fi

echo "==> Ensuring base images (postgres, migrate) exist locally"
require_image "${POSTGRES_IMAGE}" \
  "Run once on a machine with Docker Hub: bash scripts/bundle-docker-images.sh, then scp and docker load on this server."
require_image "${MIGRATE_IMAGE}" \
  "Run once on a machine with Docker Hub: bash scripts/bundle-docker-images.sh, then scp and docker load on this server."

echo "==> Ensuring database is up"
"${COMPOSE[@]}" up -d db

echo "==> Running migrations"
"${COMPOSE[@]}" up migrate --abort-on-container-exit

echo "==> Starting API"
"${COMPOSE[@]}" up -d api

if [[ "${OBSERVABILITY_ENABLED}" == "true" ]]; then
  echo "==> Ensuring observability images (prometheus, grafana) exist locally"
  require_image "${PROMETHEUS_IMAGE}" \
    "Observability images are bundled with postgres/migrate in scripts/bundle-docker-images.sh — load them on this server."
  require_image "${GRAFANA_IMAGE}" \
    "Observability images are bundled with postgres/migrate in scripts/bundle-docker-images.sh — load them on this server."

  echo "==> Starting observability stack (Prometheus + Grafana)"
  export PROMETHEUS_IMAGE GRAFANA_IMAGE
  "${COMPOSE_OBS[@]}" up -d prometheus grafana

  echo "==> Reloading Prometheus configuration"
  curl -sf -X POST http://127.0.0.1:9090/-/reload >/dev/null 2>&1 || true

  if curl -sf http://127.0.0.1:9090/-/healthy >/dev/null; then
    echo "    Prometheus: healthy (http://127.0.0.1:9090)"
  else
    echo "WARN: Prometheus health check failed — check logs: docker compose -f docker-compose.observability.yml logs prometheus" >&2
  fi

  if curl -sf http://127.0.0.1:3001/api/health >/dev/null; then
    echo "    Grafana:    healthy (http://127.0.0.1:3001)"
  else
    echo "WARN: Grafana health check failed — check logs: docker compose -f docker-compose.observability.yml logs grafana" >&2
  fi
else
  echo "==> Observability disabled (OBSERVABILITY_ENABLED=false)"
fi

echo "==> Pruning old API images (keep latest + current sha)"
docker images rahil-gallery-api --format '{{.ID}} {{.Tag}}' \
  | awk -v keep="${GIT_SHA}" '$2 != "latest" && $2 != keep {print $1}' \
  | xargs -r docker rmi -f 2>/dev/null || true

rm -f "${IMAGE_TAR}"
echo "==> Deploy complete (${GIT_SHA})"
