#!/usr/bin/env bash
# Sync Postgres password with DATABASE_URL and restart the API container.
# Use when API is crash-looping with "password authentication failed for user postgres".
#
# Usage (on VPS, from repo root):
#   bash scripts/fix-prod-db-auth.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if [[ ! -f .env ]]; then
  echo "ERROR: .env not found in ${ROOT}" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "ERROR: DATABASE_URL is not set in .env" >&2
  exit 1
fi

DB_PASSWORD="$(
  python3 - <<'PY'
from urllib.parse import urlparse
from pathlib import Path
for line in Path(".env").read_text().splitlines():
    if line.startswith("DATABASE_URL="):
        print(urlparse(line.split("=", 1)[1].strip().strip("'\"")).password or "")
        break
PY
)"

if [[ -z "${DB_PASSWORD}" ]]; then
  echo "ERROR: could not parse password from DATABASE_URL" >&2
  exit 1
fi

echo "==> Setting Postgres user password to match DATABASE_URL"
docker exec -u postgres rahil-gallery-db psql -d "${POSTGRES_DB:-rahil_gallery}" \
  -c "ALTER USER postgres WITH PASSWORD '${DB_PASSWORD}';"

# Stale DATABASE_URL_DOCKER in .env overrides the derived value and causes auth mismatch.
if grep -q '^DATABASE_URL_DOCKER=' .env; then
  echo "==> Commenting out DATABASE_URL_DOCKER in .env (will be derived from DATABASE_URL)"
  sed -i 's/^DATABASE_URL_DOCKER=/# DATABASE_URL_DOCKER=/' .env
fi

eval "$(bash scripts/export-compose-env.sh)"

echo "==> Recreating API with derived DATABASE_URL_DOCKER"
docker compose -f docker-compose.prod.yml up -d --force-recreate --no-deps api

echo "==> Waiting for API..."
for _ in $(seq 1 15); do
  if curl -fsS http://127.0.0.1:8080/health/ready >/dev/null 2>&1; then
    echo "OK: /health/ready is up"
    curl -fsS http://127.0.0.1:8080/health/ready
    echo
    exit 0
  fi
  sleep 2
done

echo "ERROR: API still not ready — check logs:" >&2
echo "  docker logs --tail=50 rahil-gallery-api" >&2
exit 1
