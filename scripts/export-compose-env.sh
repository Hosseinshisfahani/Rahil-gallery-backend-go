#!/usr/bin/env bash
# Source .env and export DATABASE_URL_DOCKER for docker compose prod commands.
# Usage: eval "$(bash scripts/export-compose-env.sh)"
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

if [[ -n "${DATABASE_URL:-}" ]]; then
  export DATABASE_URL_DOCKER="$(
    printf '%s' "$DATABASE_URL" | sed -E 's/(postgres(ql)?:\/\/[^@]+@)[^:/]+/\1db/'
  )"
elif [[ -z "${DATABASE_URL_DOCKER:-}" ]]; then
  echo "ERROR: set DATABASE_URL or DATABASE_URL_DOCKER in .env" >&2
  exit 1
fi

printf 'export DATABASE_URL_DOCKER=%q\n' "$DATABASE_URL_DOCKER"
