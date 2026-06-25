#!/usr/bin/env bash
# Package tracked server source + production CLI binaries for VPS sync.
# Usage: GITHUB_SHA=<sha> bash scripts/package-server-source.sh [output.tar.gz]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

OUT="${1:-server-source.tar.gz}"
STAGING="$(mktemp -d)"
trap 'rm -rf "${STAGING}"' EXIT

echo "==> Archiving git tree at HEAD"
git archive --format=tar HEAD | tar -x -C "${STAGING}"

mkdir -p "${STAGING}/bin"
for bin in seed-prod backfill-crm; do
  if [[ -f "bin/${bin}" ]]; then
    cp "bin/${bin}" "${STAGING}/bin/${bin}"
  else
    echo "WARN: bin/${bin} missing — run build-prod-tools in CI first" >&2
  fi
done

if [[ -n "${GITHUB_SHA:-}" ]]; then
  echo "${GITHUB_SHA}" > "${STAGING}/.deploy-sha"
fi

tar -czf "${OUT}" -C "${STAGING}" .
echo "==> Wrote ${OUT} ($(du -h "${OUT}" | awk '{print $1}'))"
