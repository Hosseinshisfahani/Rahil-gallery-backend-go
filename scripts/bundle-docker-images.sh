#!/usr/bin/env bash
# Bundle Docker images for offline transfer when Docker Hub is unreachable on the server.
# Usage:
#   bash scripts/bundle-docker-images.sh
#   scp /tmp/rahil-docker-images.tar.gz root@YOUR_SERVER:/root/
#   ssh root@YOUR_SERVER 'gunzip -c /root/rahil-docker-images.tar.gz | docker load'
set -euo pipefail

POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:16-alpine}"
MIGRATE_IMAGE="${MIGRATE_IMAGE:-migrate/migrate:v4.18.1}"
PROMETHEUS_IMAGE="${PROMETHEUS_IMAGE:-prom/prometheus:v2.55.1}"
GRAFANA_IMAGE="${GRAFANA_IMAGE:-grafana/grafana:11.4.0}"
OUT="${1:-/tmp/rahil-docker-images.tar.gz}"

echo "Pulling ${POSTGRES_IMAGE} ..."
docker pull "${POSTGRES_IMAGE}"
echo "Pulling ${MIGRATE_IMAGE} ..."
docker pull "${MIGRATE_IMAGE}"
echo "Pulling ${PROMETHEUS_IMAGE} ..."
docker pull "${PROMETHEUS_IMAGE}"
echo "Pulling ${GRAFANA_IMAGE} ..."
docker pull "${GRAFANA_IMAGE}"

echo "Saving to ${OUT} ..."
docker save "${POSTGRES_IMAGE}" "${MIGRATE_IMAGE}" "${PROMETHEUS_IMAGE}" "${GRAFANA_IMAGE}" | gzip > "${OUT}"
ls -lh "${OUT}"
echo "Done. Copy to server and run: gunzip -c ${OUT} | docker load"
