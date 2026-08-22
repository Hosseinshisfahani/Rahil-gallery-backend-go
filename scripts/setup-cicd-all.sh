#!/usr/bin/env bash
# Alias for setup-cicd.sh — configures GitHub Actions secrets for BOTH repos.
#
# Usage (from Rahil-gallery-backend-go repo root):
#   export GITHUB_TOKEN="$(github_personal_token)"
#   export REPO_DISPATCH_TOKEN="$GITHUB_TOKEN"   # optional — chains API deploy -> client deploy
#   bash scripts/setup-cicd-all.sh \
#     --host YOUR_VPS_IP \
#     --user root \
#     --server-path /root/source/Rahil-gallery-backend-go \
#     --client-path /root/source/Rahil-gallery-frontend
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec bash "${ROOT}/scripts/setup-cicd.sh" "$@"
