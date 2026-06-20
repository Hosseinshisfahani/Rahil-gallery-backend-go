# Testing — Auth

Three layers cover authentication behavior.

## 1. TDD (unit tests)

Fast tests with in-memory fakes; no database.

| Package | File | Focus |
|---------|------|--------|
| `internal/application/auth` | `service_test.go` | Register, login, refresh rotation, logout, validation |
| `internal/pkg/tokens` | `refresh_test.go` | Refresh token hashing |
| `internal/infrastructure/security` | `*_test.go` | JWT, bcrypt |

```bash
make test-unit
# or
go test ./internal/... -count=1
```

## 2. BDD (Gherkin + Godog)

Human-readable scenarios in `features/auth/authentication.feature`.

```bash
make test-bdd
# or
go test ./tests/bdd/... -count=1
```

## 3. Feature tests (HTTP)

End-to-end HTTP tests against Fiber routes with fake persistence (no Postgres).

```bash
make test-feature
# or
go test ./tests/feature/... -count=1
```

## Integration tests (optional, real Postgres)

Requires migrations applied and `TEST_DATABASE_URL` or `DATABASE_URL`.

```bash
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/rahil_gallery?sslmode=disable"
make test-integration
```

## Run everything

```bash
go mod tidy
make test
```

## Dev database seed

Populates staff users, customers (CRM profiles with age group and gender), catalog products, orders, and wishlist items for local admin UI testing. Default scale: **10,000 customers** (8 fixed fixtures + bulk generated rows). Supports **8–100,000** via `--customers`.

```bash
make docker-dev && make migrate-up
make seed                    # 10,000 customers (default)
SEED_CUSTOMERS=50000 make seed-reset   # wipe and insert 50k
make seed-small              # 8 fixture customers only (fast)
go run ./cmd/seed --customers=100000 --reset
```

Bulk customers use phones `+98900XXXXXXXX` and emails `seed-*@rehil.dev` so reset stays fast. Credentials are printed on success. Refuses to run when `APP_ENV=production`.

## Production bootstrap

Creates only what production needs after migrations — an initial **admin** account (and optional **staff**). No demo customers, catalog, or orders.

```bash
# On VPS after migrate (set strong password in .env first)
export APP_ENV=production
export SEED_ADMIN_PASSWORD='YourSecurePass1'
make seed-prod

# Rotate admin password later
export SEED_ADMIN_PASSWORD='NewSecurePass1'
make seed-prod-password
```

Environment variables:

| Variable | Required | Default |
|----------|----------|---------|
| `SEED_ADMIN_PASSWORD` | yes | — |
| `SEED_ADMIN_EMAIL` | no | `admin@rehil.gallery` |
| `SEED_ADMIN_FIRST_NAME` | no | `Admin` |
| `SEED_ADMIN_LAST_NAME` | no | `Rehil` |
| `SEED_ADMIN_PHONE` | no | — |
| `SEED_STAFF_EMAIL` | no* | — |
| `SEED_STAFF_PASSWORD` | no* | — |

\* Both staff email and password must be set to create a staff account.

Password rules match registration: at least 8 characters with one letter and one digit. Passwords are never printed. Idempotent: skips if an admin account already exists.
