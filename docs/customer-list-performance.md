# Customer List Query Optimization

Migrations: `000004`–`000008` (commerce stats, quick search, phone digits, advanced indexes, stored segment)

## Problem

`GET /api/v1/admin/customers` was slow (~650–800 ms with 10k seeded customers) because each list request:

1. Ran **two** heavy queries (`COUNT` + `SELECT`)
2. Used a **`LATERAL` join** aggregating `orders` per customer row
3. Used **`ILIKE '%…%'`** on computed columns without usable indexes

## Solution

### 1. Denormalized commerce stats on `customer_profiles`

| Column | Source |
|--------|--------|
| `total_orders` | `COUNT(orders)` |
| `total_ltv` | `SUM(orders.total_amount)` |
| `last_purchase_at` | `MAX(orders.placed_at)` |
| `first_purchase_at` | `MIN(orders.placed_at)` |

- **Backfilled** on migration
- **Kept in sync** via trigger `trg_orders_sync_commerce_stats` on `orders`
- **Bulk seed** disables trigger during COPY, then runs `refresh_all_customer_commerce_stats()`

List queries read `cp.total_*` instead of aggregating `orders` per row.

### 2. Trigram search indexes (`pg_trgm`)

- `idx_users_full_name_trgm` — name search
- `idx_users_phone_trgm` — phone search

### 3. Quick search vs advanced filters (separate code paths)

| Mode | Query params | SQL path |
|------|--------------|----------|
| **Quick** | `q` only | Name + phone (`buildQuickSearchWhere`) |
| **Advanced** | `id`, `email`, `segment`, LTV, dates, tags, … | Structured filters (`buildAdvancedWhere`) |

When any advanced param is present, `q` is ignored (handler enforces mutual exclusion).

**Quick search (`q`)** — index-friendly paths (no `OR` full-table scan)

| Input | Strategy | Index |
|-------|----------|-------|
| Text only (`ali`) | Name `LIKE '%…%'` only | `idx_users_full_name_trgm` |
| 4+ digits only (`98900`) | `phone_digits LIKE 'digits%'` | `idx_users_phone_digits_prefix` |
| Mixed (`ali98900`) | `UNION` of name + phone subqueries | Both indexes |

Quick-search requests run `COUNT` and `SELECT` **in parallel** (two pool connections).

**Advanced search** — profile-first `MATERIALIZED` CTE when filtering on `customer_profiles`

| Param | Strategy | Index |
|-------|----------|-------|
| `id` | `u.id = $1` (users-first path) | `users_pkey` |
| `email` | `lower(email) LIKE` (users-first) | `idx_users_email_trgm` |
| `status`, `registeredFrom/To` | users-first | `idx_users_status`, `idx_users_customer_created` |
| `segment` | `segment = $1` on stored column | `idx_customer_profiles_segment` |
| `vip`, `ltv*`, `orders*`, dates on profile, `tags`, `hasPurchased` | `profile_filter` CTE → join users | `idx_customer_profiles_*` |

Problem (before): `users LEFT JOIN customer_profiles` scanned 100k users, then filtered on `cp.*` (~350–520 ms per query).

Fix: `WITH profile_filter AS MATERIALIZED (SELECT user_id FROM customer_profiles WHERE …)` drives from indexed profile columns, then joins matching users.

Advanced requests also run `COUNT` + `SELECT` in parallel.

## Apply

```bash
make migrate-up
# If already seeded:
make seed   # with --reset if needed, or:
# docker compose exec db psql -U postgres -d rahil_gallery -c "SELECT refresh_all_customer_commerce_stats();"
```

## Verify

```bash
# Compare before/after with EXPLAIN ANALYZE in psql
docker compose exec db psql -U postgres -d rahil_gallery -c "
EXPLAIN ANALYZE
SELECT COUNT(*)
FROM users u
INNER JOIN roles r ON r.id = u.role_id AND r.name = 'customer'
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
WHERE u.deleted_at IS NULL
  AND replace(COALESCE(u.phone, ''), ' ', '') LIKE '98900%';
"
```

## Expected improvement

| Scenario | Before (approx) | After (target) |
|----------|-----------------|----------------|
| Text quick search `q=ali` (100k rows) | 550–700 ms | 30–80 ms |
| Phone prefix `q=98900` | 550–700 ms | 20–50 ms |
| List page 1, no search | 400–700 ms | 30–80 ms |
| UUID lookup (advanced `id=`) | 650–800 ms | &lt; 10 ms |

Actual numbers depend on hardware, Docker, and cache warmth.

### 4. Stored segment column (migration `000008`)

- `customer_profiles.segment` — maintained by trigger + commerce-stats refresh
- List/detail read `cp.segment` instead of runtime `CASE`
- Segment filter: `WHERE segment = 'returning'` (btree index)

### 5. Optional skip count (`includeTotal=false`)

Skips `COUNT(*)`; returns `meta.hasMore` via `LIMIT perPage+1`:

```bash
GET /api/v1/admin/customers?q=ali&includeTotal=false&page=1&perPage=10
```

Response `meta` omits `total` / `totalPages`, includes `hasMore: true|false`.

## Future optimizations

- Short-TTL cache for repeated search queries
- Generated columns `full_name_norm` for simpler name indexes
- Read replica for admin reporting
