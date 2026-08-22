#!/usr/bin/env bash
# Build a Linux API binary on a machine with Go module access, for offline VPS deploy.
# Usage:
#   bash scripts/bundle-api-binary.sh
#   scp /tmp/rahil-api-linux.tar.gz root@YOUR_SERVER:/tmp/
#   ssh root@YOUR_SERVER 'mkdir -p ~/source/Rahil-gallery-backend-go/bin && tar -xzf /tmp/rahil-api-linux.tar.gz -C ~/source/Rahil-gallery-backend-go && chmod +x ~/source/Rahil-gallery-backend-go/bin/rahil-api'
#   ssh root@YOUR_SERVER 'cd ~/source/Rahil-gallery-backend-go && make docker-dev && ./bin/rahil-api'
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-/tmp/rahil-api-linux.tar.gz}"
BIN="${ROOT}/bin/rahil-api"

cd "${ROOT}"
mkdir -p bin

export GOPROXY="${GOPROXY:-https://goproxy.io,https://goproxy.cn,direct}"

echo "==> Building static Linux amd64 binary (GOPROXY=${GOPROXY})"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "${BIN}" ./cmd/api

echo "==> Packing ${OUT}"
tar -czf "${OUT}" -C "${ROOT}" bin/rahil-api
ls -lh "${OUT}"
echo "Done. Copy to VPS and run ./bin/rahil-api (with make docker-dev for Postgres)."
