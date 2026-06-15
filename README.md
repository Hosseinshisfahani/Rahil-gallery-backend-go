# Rahil Gallery Server

Jewelry e-commerce API — Go, [Fiber](https://gofiber.io/), DDD / Clean Architecture.

**Full technical workflow:** [docs/technical-workflow.md](docs/technical-workflow.md) — architecture, auth flow, Docker, DB, testing, API reference.

## Project layout

```
├── cmd/api/                 # HTTP entrypoint (Fiber)
├── internal/
│   ├── domain/              # Entities & domain types (per bounded context)
│   ├── application/         # Use cases / services
│   ├── infrastructure/      # DB, cache, external gateways
│   └── interfaces/          # HTTP handlers, DTOs, middleware
├── migrations/              # PostgreSQL schema
└── docs/                    # Technical docs (workflow, schema, testing)
```

## Database schema

Initial migration: `migrations/000001_init_schema.up.sql`

Full ER overview and context map: [docs/database-schema.md](docs/database-schema.md)

Customer list performance: [docs/customer-list-performance.md](docs/customer-list-performance.md)

```bash
# Example: apply schema (Docker Postgres + migrate image)
make docker-dev
make migrate-up
```

## Domain packages

| Context   | Package                      |
|-----------|------------------------------|
| Identity  | `internal/domain/identity`   |
| Catalog   | `internal/domain/catalog`    |
| Inventory | `internal/domain/inventory`  |
| Cart      | `internal/domain/cart`       |
| Order     | `internal/domain/order`      |
| Payment   | `internal/domain/payment`    |
| Promotion | `internal/domain/promotion`  |
| Review    | `internal/domain/review`     |
| Wishlist  | `internal/domain/wishlist`   |

## Docker Compose

**Full stack** (Postgres + migrations + API):

```bash
cp .env.example .env
docker compose up -d --build
# API: http://localhost:8080
```

**DB only** (run API on host with `go run`):

```bash
make docker-dev
make run
# API: http://localhost:8080
```

`make run` loads `.env` automatically (including `DATABASE_URL` on port 55432).

Useful commands:

```bash
make docker-up      # build & start all services
make docker-down    # stop and remove containers
make docker-logs    # follow API logs
make docker-migrate # re-run migrations
```

**Docker build fails with `403` on `proxy.golang.org` (VPN):** use an alternate proxy or vendor modules:

```bash
# Option A: alternate GOPROXY (try with VPN on or off)
GOPROXY=https://goproxy.io,direct docker compose build api
GOPROXY=https://goproxy.cn,direct docker compose build api

# Option B: vendor on host (while go modules download works), then offline Docker build
make docker-up-vendor
```

## Run the API (without Docker)

```bash
cp .env.example .env
# edit DATABASE_URL if needed

go mod tidy
go run ./cmd/api
```

Endpoints:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness |
| GET | `/health/ready` | DB readiness |
| GET | `/api/v1/` | API info |

### Auth (JWT)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | — | Register customer account |
| POST | `/api/v1/auth/login` | — | Login, returns access + refresh tokens |
| POST | `/api/v1/auth/refresh` | — | Rotate refresh token |
| POST | `/api/v1/auth/logout` | — | Revoke refresh token (body: `refresh_token`) |
| GET | `/api/v1/auth/me` | Bearer JWT | Current user profile |

Password rules: min 8 chars, at least one letter and one digit.

```bash
# Register
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Secret12","first_name":"Ali","last_name":"Rahil"}'

# Login
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Secret12"}'

# Me (replace TOKEN)
curl -s http://localhost:8080/api/v1/auth/me -H "Authorization: Bearer TOKEN"
```

## Repository ports (domain)

Interfaces live next to each context, e.g. `internal/domain/catalog/repository.go`.  
Postgres implementations go under `internal/infrastructure/persistence/postgres/`.

## Testing

See [docs/testing.md](docs/testing.md).

```bash
go mod tidy
make test          # all packages (unit + bdd + feature)
make test-unit     # TDD unit tests
make test-bdd      # Gherkin scenarios
make test-feature  # HTTP feature tests
```

## Next steps

- Implement Postgres repositories per port
- Application use cases (auth, catalog, cart, checkout)
- JWT middleware and route groups under `/api/v1`
