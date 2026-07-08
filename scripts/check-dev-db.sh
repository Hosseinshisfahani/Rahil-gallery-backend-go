#!/usr/bin/env bash
# Verify dev Postgres is reachable before starting the API.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [ -f .env ]; then
	set -a
	# shellcheck disable=SC1091
	source .env
	set +a
fi

POSTGRES_PORT="${POSTGRES_PORT:-5433}"

db_container_running() {
	docker inspect rahil-gallery-db >/dev/null 2>&1
}

check_via_container() {
	docker exec rahil-gallery-db pg_isready -U postgres -d rahil_gallery >/dev/null 2>&1
}

check_via_host_port() {
	docker run --rm --add-host=host.docker.internal:host-gateway \
		-e "PGPASSWORD=${POSTGRES_PASSWORD:-postgres}" \
		"postgres:16-alpine" \
		psql -h host.docker.internal -p "${POSTGRES_PORT}" -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-rahil_gallery}" -c 'SELECT 1' >/dev/null 2>&1
}

echo "check-dev-db: Rahil Gallery Postgres (port ${POSTGRES_PORT})"

if db_container_running && check_via_container; then
	echo "check-dev-db: ok (rahil-gallery-db)"
	exit 0
fi

if check_via_host_port 2>/dev/null; then
	echo "check-dev-db: ok (127.0.0.1:${POSTGRES_PORT})"
	exit 0
fi

# Fallback: local go dbping when go is available
HOST_DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@127.0.0.1:${POSTGRES_PORT}/rahil_gallery?sslmode=disable&connect_timeout=3}"
if command -v go >/dev/null 2>&1 && DATABASE_URL="$HOST_DATABASE_URL" go run ./scripts/devtools/dbping/main.go 2>/dev/null; then
	echo "check-dev-db: ok (DATABASE_URL)"
	exit 0
fi

echo "check-dev-db: cannot connect to Rahil Gallery Postgres" >&2
echo >&2

if ! db_container_running; then
	echo "  rahil-gallery-db is not running." >&2
	echo "  Fix: make docker-dev" >&2
else
	echo "  Container is up but DB is not ready yet — wait a few seconds and retry." >&2
	echo "  Or run: make docker-dev && make migrate-up" >&2
fi

if timeout 1 bash -c 'echo > /dev/tcp/127.0.0.1/5432' 2>/dev/null && [ "${POSTGRES_PORT}" != "5432" ]; then
	echo >&2
	echo "  Note: port 5432 is in use by another process (e.g. a different Postgres)." >&2
	echo "  Rahil Gallery uses port ${POSTGRES_PORT} — set DATABASE_URL in .env accordingly." >&2
fi

exit 1
