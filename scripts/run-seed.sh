#!/usr/bin/env bash
# Run dev seed. Prefers host DATABASE_URL; falls back to Compose network (db:5432).
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
SEED_BIN="$ROOT/bin/seed-dev"
HOST_DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@127.0.0.1:${POSTGRES_PORT}/rahil_gallery?sslmode=disable&connect_timeout=5}"
DOCKER_DATABASE_URL="${SEED_DATABASE_URL:-postgres://postgres:postgres@db:5432/rahil_gallery?sslmode=disable}"
ARGS=("$@")

build_seed() {
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$SEED_BIN" ./cmd/seed
}

run_host() {
	DATABASE_URL="$HOST_DATABASE_URL" go run ./cmd/seed "${ARGS[@]}"
}

run_docker() {
	local network
	network="$(docker inspect rahil-gallery-db --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' 2>/dev/null || true)"
	if [ -z "$network" ]; then
		echo "run-seed: rahil-gallery-db container not found — run: make docker-dev" >&2
		return 1
	fi
	build_seed
	docker run --rm \
		--network "$network" \
		-e "APP_ENV=${APP_ENV:-development}" \
		-e "DATABASE_URL=${DOCKER_DATABASE_URL}" \
		-v "$SEED_BIN:/seed:ro" \
		alpine:3.20 /seed "${ARGS[@]}"
}

if command -v go >/dev/null 2>&1 && DATABASE_URL="$HOST_DATABASE_URL" go run ./scripts/devtools/dbping/main.go 2>/dev/null; then
	run_host
elif docker inspect rahil-gallery-db >/dev/null 2>&1; then
	run_docker
else
	echo "run-seed: Postgres not reachable — run: make docker-dev" >&2
	exit 1
fi
