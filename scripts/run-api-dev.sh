#!/usr/bin/env bash
# Start/stop the dev API on APP_PORT (default :8081). Requires Rahil Gallery Postgres (make docker-dev).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

CONTAINER_NAME="${API_DEV_CONTAINER:-rahil-gallery-api-dev}"
API_BIN="$ROOT/bin/rahil-api-dev"
API_DATABASE_URL="${API_DATABASE_URL:-postgres://postgres:postgres@db:5432/rahil_gallery?sslmode=disable}"
PID_FILE="$ROOT/.api-dev.pid"

if [ -f .env ]; then
	set -a
	# shellcheck disable=SC1091
	source .env
	set +a
fi

build_api() {
	mkdir -p bin data/customer-signatures
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$API_BIN" ./cmd/api
}

run_docker() {
	local network="$1"
	build_api
	docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
	docker run -d --name "$CONTAINER_NAME" \
		--network "$network" \
		-p "${APP_PORT:-8081}:8080" \
		-e "APP_ENV=${APP_ENV:-development}" \
		-e "APP_HOST=0.0.0.0" \
		-e "APP_PORT=8080" \
		-e "DATABASE_URL=${API_DATABASE_URL}" \
		-e "JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET:-dev-docker-secret-change-in-production}" \
		-e "JWT_ACCESS_TTL=${JWT_ACCESS_TTL:-15m}" \
		-e "JWT_REFRESH_TTL=${JWT_REFRESH_TTL:-168h}" \
		-e "CATALOG_ASSETS_DIR=/app/data/catalog-images" \
		-e "CUSTOMER_SIGNATURES_DIR=/app/data/customer-signatures" \
		-v "$API_BIN:/app/api:ro" \
		-v "$ROOT/data/catalog-images:/app/data/catalog-images:ro" \
		-v "$ROOT/data/customer-signatures:/app/data/customer-signatures" \
		alpine:3.20 /app/api >/dev/null
	if ! docker ps --filter "name=^${CONTAINER_NAME}$" --filter status=running -q | grep -q .; then
		echo "run-api-dev: failed to start API container (port ${APP_PORT:-8081} in use?)" >&2
		docker logs "$CONTAINER_NAME" 2>&1 | tail -15 >&2 || true
		docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
		exit 1
	fi
	echo "run-api-dev: API listening on http://localhost:${APP_PORT:-8081} (container: $CONTAINER_NAME)"
}

run_host_foreground() {
	bash scripts/check-dev-db.sh
	exec go run ./cmd/api
}

run_host_background() {
	bash scripts/check-dev-db.sh
	mkdir -p data/customer-signatures
	if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
		echo "run-api-dev: API already running (pid $(cat "$PID_FILE"))"
		exit 0
	fi
	nohup go run ./cmd/api >"$ROOT/.api-dev.log" 2>&1 &
	echo $! >"$PID_FILE"
	echo "run-api-dev: API listening on http://localhost:${APP_PORT:-8081} (pid $(cat "$PID_FILE"))"
}

stop_api() {
	if [ -f "$PID_FILE" ]; then
		pid="$(cat "$PID_FILE")"
		if kill -0 "$pid" 2>/dev/null; then
			kill "$pid" 2>/dev/null || true
			echo "run-api-dev: stopped host API (pid $pid)"
		fi
		rm -f "$PID_FILE"
	fi
	if docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1; then
		echo "run-api-dev: stopped $CONTAINER_NAME"
	fi
}

use_docker_api() {
	local network
	network="$(docker inspect rahil-gallery-db --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' 2>/dev/null || true)"
	[ -n "$network" ] || return 1
	# Host DATABASE_URL unreachable but container DB is up — run API on Compose network.
	if ! command -v go >/dev/null 2>&1; then
		run_docker "$network"
		return 0
	fi
	if [ -f .env ] && [ -n "${DATABASE_URL:-}" ]; then
		if ! DATABASE_URL="$DATABASE_URL" go run ./scripts/devtools/dbping/main.go 2>/dev/null; then
			run_docker "$network"
			return 0
		fi
	fi
	return 1
}

case "${1:-start}" in
	stop)
		stop_api
		;;
	start)
		if use_docker_api; then
			exit 0
		fi
		run_host_background
		;;
	run)
		if use_docker_api; then
			docker logs -f "$CONTAINER_NAME"
			exit 0
		fi
		run_host_foreground
		;;
	*)
		echo "usage: $0 [start|stop|run]" >&2
		exit 1
		;;
esac
