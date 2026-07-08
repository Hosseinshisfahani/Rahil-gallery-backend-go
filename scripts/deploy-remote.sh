#!/usr/bin/env bash
# Server-side deploy hook — called by GitHub Actions over SSH.
# Usage: bash scripts/deploy-remote.sh /tmp/api-image.tar.gz <git-sha> [/tmp/base-images.tar.gz]
set -euo pipefail

IMAGE_TAR="${1:?image tar path required}"
GIT_SHA="${2:-latest}"
BASE_IMAGES_TAR="${3:-/tmp/base-images.tar.gz}"
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

missing_image() {
  local image="$1"
  ! docker image inspect "${image}" >/dev/null 2>&1
}

needs_bundled_images() {
  if missing_image "${POSTGRES_IMAGE}" || missing_image "${MIGRATE_IMAGE}"; then
    return 0
  fi
  if [[ "${OBSERVABILITY_ENABLED}" == "true" ]]; then
    if missing_image "${PROMETHEUS_IMAGE}" || missing_image "${GRAFANA_IMAGE}"; then
      return 0
    fi
  fi
  return 1
}

load_bundled_images() {
  if [[ ! -f "${BASE_IMAGES_TAR}" ]]; then
    return 1
  fi

  echo "==> Loading bundled base images from ${BASE_IMAGES_TAR}"
  gunzip -c "${BASE_IMAGES_TAR}" | docker load
  rm -f "${BASE_IMAGES_TAR}"
  return 0
}

ensure_bundled_images() {
  if ! needs_bundled_images; then
    return
  fi

  if load_bundled_images && ! needs_bundled_images; then
    echo "==> Bundled base images loaded successfully"
    return
  fi

  echo "ERROR: required Docker base images are missing on this server." >&2
  echo "       CI should upload ${BASE_IMAGES_TAR} during deploy." >&2
  echo "       Manual fallback: bash scripts/bundle-docker-images.sh locally," >&2
  echo "       scp the tarball to the server, then docker load." >&2
  exit 1
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

database_url_component() {
  local component="$1"
  local url="$2"
  python3 - "${component}" "${url}" <<'PY'
import sys
from urllib.parse import unquote, urlparse

component = sys.argv[1]
parsed = urlparse(sys.argv[2])

if component == "username":
    print(unquote(parsed.username or ""))
elif component == "password":
    print(unquote(parsed.password or ""))
elif component == "database":
    print(unquote((parsed.path or "/").lstrip("/")))
else:
    raise SystemExit(f"unknown component: {component}")
PY
}

# Migrate/API use host `db` inside Compose. Always derive from DATABASE_URL when set
# so a stale DATABASE_URL_DOCKER in .env cannot break deploy.
prepare_database_url_docker() {
  if [[ -n "${DATABASE_URL:-}" ]]; then
    export DATABASE_URL_DOCKER="$(
      printf '%s' "$DATABASE_URL" | sed -E 's/(postgres(ql)?:\/\/[^@]+@)[^:/]+/\1db/'
    )"
    export POSTGRES_PASSWORD="$(database_url_component password "${DATABASE_URL}")"
    echo "==> Using DATABASE_URL_DOCKER derived from DATABASE_URL"
    echo "==> Using POSTGRES_PASSWORD derived from DATABASE_URL"
    return
  fi
  if [[ -n "${DATABASE_URL_DOCKER:-}" ]]; then
    export DATABASE_URL_DOCKER
    export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-$(database_url_component password "${DATABASE_URL_DOCKER}")}"
    echo "==> Using DATABASE_URL_DOCKER from .env"
    echo "==> Using POSTGRES_PASSWORD derived from DATABASE_URL_DOCKER"
    return
  fi
  echo "ERROR: set DATABASE_URL (recommended) or DATABASE_URL_DOCKER in .env" >&2
  exit 1
}

validate_postgres_credentials() {
  local source_url="${DATABASE_URL:-${DATABASE_URL_DOCKER:-}}"
  local db_user db_password db_name db_network

  if [[ -z "${source_url}" ]]; then
    echo "ERROR: database URL is required before validating Postgres connectivity" >&2
    exit 1
  fi

  db_user="$(database_url_component username "${source_url}")"
  db_password="$(database_url_component password "${source_url}")"
  db_name="$(database_url_component database "${source_url}")"

  if [[ -z "${db_user}" || -z "${db_password}" || -z "${db_name}" ]]; then
    echo "ERROR: database URL must include username, password, and database name" >&2
    exit 1
  fi

  db_network="$(
    docker inspect rahil-gallery-db \
      --format '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' \
      | awk 'NF {print; exit}'
  )"

  if [[ -z "${db_network}" ]]; then
    echo "ERROR: could not determine Docker network for rahil-gallery-db" >&2
    exit 1
  fi

  echo "==> Validating Postgres credentials from Docker network"
  if ! docker run --rm \
    --network "${db_network}" \
    -e "PGPASSWORD=${db_password}" \
    "${POSTGRES_IMAGE}" \
    psql -v ON_ERROR_STOP=1 -h db -U "${db_user}" -d "${db_name}" -c 'select 1' >/dev/null; then
    cat >&2 <<'MSG'
ERROR: Postgres rejected the credentials derived from .env DATABASE_URL.

Fix the server .env so DATABASE_URL matches the existing Postgres role password,
or rotate the Postgres role password as a one-time operational step. The deploy
will not rewrite database credentials automatically.
MSG
    exit 1
  fi
}

load_env
prepare_database_url_docker

echo "==> Loading API image from ${IMAGE_TAR}"
gunzip -c "${IMAGE_TAR}" | docker load

if docker image inspect "rahil-gallery-api:${GIT_SHA}" >/dev/null 2>&1; then
  docker tag "rahil-gallery-api:${GIT_SHA}" rahil-gallery-api:latest
  echo "==> Tagged rahil-gallery-api:${GIT_SHA} as latest"
else
  LOADED="$(docker images --format '{{.Repository}}:{{.Tag}}' | grep "^rahil-gallery-api:${GIT_SHA}$" | head -1 || true)"
  if [[ -z "${LOADED}" ]]; then
    LOADED="$(docker images --format '{{.Repository}}:{{.Tag}}' | grep '^rahil-gallery-api:' | grep -v ':latest$' | head -1 || true)"
  fi
  if [[ -z "${LOADED}" ]]; then
    echo "ERROR: rahil-gallery-api:${GIT_SHA} not found after docker load" >&2
    exit 1
  fi
  docker tag "${LOADED}" rahil-gallery-api:latest
  echo "==> Tagged ${LOADED} as latest"
fi

export API_IMAGE="rahil-gallery-api:latest"

if [[ ! -f .env ]]; then
  echo "WARN: .env missing — copy from .env.example and set JWT_ACCESS_SECRET + DATABASE_URL" >&2
fi

echo "==> Ensuring base images (postgres, migrate) exist locally"
ensure_bundled_images
require_image "${POSTGRES_IMAGE}" \
  "Bundled base images were not loaded. Re-run deploy from CI or load scripts/bundle-docker-images.sh output manually."
require_image "${MIGRATE_IMAGE}" \
  "Bundled base images were not loaded. Re-run deploy from CI or load scripts/bundle-docker-images.sh output manually."

echo "==> Ensuring database is up"
"${COMPOSE[@]}" up -d db

validate_postgres_credentials

echo "==> Running migrations"
"${COMPOSE[@]}" up migrate --abort-on-container-exit

echo "==> Starting API (force recreate so :latest image updates apply)"
"${COMPOSE[@]}" up -d --force-recreate --no-deps api

RUNNING_IMAGE="$(docker inspect rahil-gallery-api --format '{{.Image}}' 2>/dev/null || true)"
LATEST_IMAGE="$(docker image inspect rahil-gallery-api:latest --format '{{.Id}}' 2>/dev/null || true)"
if [[ -n "${RUNNING_IMAGE}" && -n "${LATEST_IMAGE}" && "${RUNNING_IMAGE}" != "${LATEST_IMAGE}" ]]; then
  echo "ERROR: API container is not running the latest loaded image" >&2
  echo "       container=${RUNNING_IMAGE}" >&2
  echo "       latest=${LATEST_IMAGE}" >&2
  exit 1
fi
echo "    API container image matches rahil-gallery-api:latest (${GIT_SHA})"

if [[ "${OBSERVABILITY_ENABLED}" == "true" ]]; then
  echo "==> Ensuring observability images (prometheus, grafana) exist locally"
  require_image "${PROMETHEUS_IMAGE}" \
    "Bundled base images were not loaded. Re-run deploy from CI or load scripts/bundle-docker-images.sh output manually."
  require_image "${GRAFANA_IMAGE}" \
    "Bundled base images were not loaded. Re-run deploy from CI or load scripts/bundle-docker-images.sh output manually."

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
if [[ -f .deploy-sha ]]; then
  echo "==> Deploy complete (source + image at $(cat .deploy-sha))"
else
  echo "==> Deploy complete (${GIT_SHA})"
fi
