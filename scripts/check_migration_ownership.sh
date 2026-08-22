#!/usr/bin/env bash
# Fail if a NEW migration file mutates tables owned by the other stack.
# Historical Go migrations through 000014 are grandfathered (commerce scaffold + CRM era).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:-}"

# Go freeze watermark: files numbered above this must not touch Django-owned commerce tables.
GO_FREEZE_AFTER=14

die() { echo "ERROR: $*" >&2; exit 1; }

migration_seq() {
  local base
  base="$(basename "$1")"
  # 000004_name.up.sql → 4
  echo "$base" | sed -E 's/^0*([0-9]+).*/\1/'
}

if [[ "$MODE" == "go" ]]; then
  if [[ -d "$ROOT/migrations" ]]; then
    DIR="${2:-$ROOT/migrations}"
  else
    DIR="${2:-$ROOT/../Rahil-gallery-backend-go/migrations}"
  fi
  PATTERN='\b(carts|cart_items|orders|order_items|order_status_history|payments|coupons|coupon_redemptions)\b|\b(cart_status|order_status|payment_status|payment_provider|discount_type)\b'
  bad=0
  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    seq="$(migration_seq "$file")"
    if [[ "$seq" =~ ^[0-9]+$ ]] && (( seq > GO_FREEZE_AFTER )); then
      echo "Blocked (post-freeze): $file"
      bad=1
    fi
  done < <(rg -l -i -e "$PATTERN" "$DIR" --glob '*.sql' 2>/dev/null || true)
  if (( bad )); then
    die "Go migrations after 000${GO_FREEZE_AFTER} must not reference Django-owned commerce tables (see docs/strangler-fig-ownership.md)"
  fi
  echo "OK: Go migration ownership check passed ($DIR; freeze after 000${GO_FREEZE_AFTER})"
  exit 0
fi

if [[ "$MODE" == "django" ]]; then
  if [[ -f "$ROOT/manage.py" ]]; then
    DIR="${2:-$ROOT}"
  else
    DIR="${2:-$ROOT/../Rahil-gallery-backend-django}"
  fi
  PATTERN="db_table=['\"]?(users|roles|refresh_tokens|user_addresses|customers|products|product_variants|categories|collections|inventory_items|observability_events)"
  if rg -n -e "migrations\.(AddField|AlterField|RemoveField|RenameField|DeleteModel)" \
      --glob '*/migrations/*.py' "$DIR" 2>/dev/null | rg -i 'users|roles|refresh_tokens|customers|products|product_variants' >/dev/null 2>&1; then
    die "Django migrations must not alter Go-owned tables (see docs/adr-table-ownership.md)"
  fi
  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    if rg -q "managed = False" "$file"; then
      continue
    fi
    if rg -q "$PATTERN" "$file"; then
      die "Suspicious Go-table reference without managed=False in $file"
    fi
  done < <(rg -l -e "$PATTERN" --glob '*/migrations/*.py' "$DIR" 2>/dev/null || true)
  echo "OK: Django migration ownership check passed ($DIR)"
  exit 0
fi

die "Usage: $0 go|django [path]"
