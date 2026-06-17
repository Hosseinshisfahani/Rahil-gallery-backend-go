#!/usr/bin/env bash
# Bundle Docker images for offline transfer when Docker Hub is unreachable on the server.
# Usage:
#   bash scripts/bundle-docker-images.sh
#   scp /tmp/rahil-docker-images.tar.gz root@YOUR_SERVER:/root/
#   ssh root@YOUR_SERVER 'gunzip -c /root/rahil-docker-images.tar.gz | docker load'
set -euo pipefail

POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:16-alpine}"
MIGRATE_IMAGE="${MIGRATE_IMAGE:-migrate/migrate:v4.18.1}"
OUT="${1:-/tmp/rahil-docker-images.tar.gz}"

echo "Pulling ${POSTGRES_IMAGE} ..."
docker pull "${POSTGRES_IMAGE}"
echo "Pulling ${MIGRATE_IMAGE} ..."
docker pull "${MIGRATE_IMAGE}"

echo "Saving to ${OUT} ..."
docker save "${POSTGRES_IMAGE}" "${MIGRATE_IMAGE}" | gzip > "${OUT}"
ls -lh "${OUT}"
echo "Done. Copy to server and run: gunzip -c ${OUT} | docker load"
