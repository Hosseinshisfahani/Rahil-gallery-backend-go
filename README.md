# Rahil Gallery Backend (Go)

Jewelry e-commerce API — Go and [Fiber](https://gofiber.io/).

This service owns **auth**, **CRM customers**, and **SMS**. Both backends share the `rahil_gallery` PostgreSQL database and the same `JWT_ACCESS_SECRET`.

The Next.js app in [Rahil-gallery-frontend](../Rahil-gallery-frontend) proxies `/api/v1` to this process (default `:8081`) except commerce prefixes, which go to Django.

## Local development (recommended)

Postgres on host `:5433`, API on host `:8081` — matches the frontend `.env.example`.

```bash
cp .env.example .env   # set JWT_ACCESS_SECRET (share with Django)
make docker-dev        # Postgres + golang-migrate
make dev               # API on :8081
```

Health: `GET http://localhost:8081/health`

Optional seed data (after migrations):

```bash
make seed-small
```

## SMS (Kavenegar)

Bulk send and birthday Lookup are Go-owned. Local default is the noop provider until `KAVENEGAR_ENABLED=true`.

| Env | Notes |
|-----|--------|
| `KAVENEGAR_API_KEY` | Panel API key |
| `KAVENEGAR_SENDER` / `SMS_SENDER_ID` | Sender line — must be authorized for that key |
| `KAVENEGAR_BIRTHDAY_TEMPLATE` | Approved Lookup template name (default `birthday`) |
| `KAVENEGAR_ENABLED` | `false` → noop; `true` → live API |
| `SMS_BIRTHDAY_CRON` | Quote the cron expr (e.g. `"0 9 * * *"`) |
| `SMS_BULK_BATCH_SIZE` | Max 200 (Kavenegar send limit) |

**Live QA tip:** If bulk jobs fail with provider status **427** (`استفاده از این خط نیازمند ایجاد سطح دسترسی می باشد`), grant access for the configured sender line on that API key in the Kavenegar console, or set `KAVENEGAR_SENDER` to a line already allowed. Failed job details appear as `lastError` on `GET /api/v1/admin/sms/jobs` and on `/admin/sms`.

**Filter tip:** Bulk SMS uses the same customer filters as the list. Quick search alone must not send `gender: "all"` as an advanced filter (fixed in client + server).

## Project layout

```
├── cmd/api/                 # HTTP entrypoint (Fiber)
├── internal/
│   ├── config/              # Environment configuration
│   ├── handler/             # HTTP routes, DTOs, middleware
│   ├── service/             # Auth, customers, SMS
│   ├── repository/          # Postgres access
│   ├── model/               # Entities and shared errors
│   └── seed/                # Local and production seed data
├── migrations/              # golang-migrate (Go-owned schema)
└── docs/
```

Do not add new Go migrations that `CREATE` / `ALTER` / `DROP` Django-owned commerce tables. `make check-ownership` enforces this.

## Configuration

See `.env.example`.

| Variable | Default | Purpose |
|----------|---------|---------|
| `APP_PORT` | `8081` | Listen port for `make dev` |
| `POSTGRES_PORT` | `5433` | Host port for `make docker-dev` |
| `DATABASE_URL` | `postgres://postgres:postgres@127.0.0.1:5433/rahil_gallery?sslmode=disable` | API database |
| `JWT_ACCESS_SECRET` | — | HS256 secret; **must match Django** |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime |

## Docker Compose

**DB only** (API on the host with `make dev`) is the usual local setup, above.

**Full stack** (Postgres + migrations + API in Compose). Compose publishes the API on `${APP_PORT:-8080}`:

```bash
cp .env.example .env
docker compose up -d --build
# API: http://localhost:8080  (or APP_PORT from .env)
```

```bash
make docker-up       # build & start all services
make docker-down     # stop and remove containers
make docker-logs     # follow API logs
make docker-migrate  # re-run migrations
```

If `docker compose build` fails with `403` on `proxy.golang.org`:

```bash
GOPROXY=https://goproxy.io,direct docker compose build api
# or vendor on the host, then:
make docker-up-vendor
```

## API surface

| Area | Status | Docs |
|------|--------|------|
| Auth (register, login, refresh, logout, me) | Live | below |
| Admin customers CRM | Live | [docs/admin-customers-api.md](docs/admin-customers-api.md) |
| Admin SMS | Live | above |

Prefix: `/api/v1`. Envelope: `{ "success": true, "data": … }` or `{ "success": false, "error": { "code", "message" } }`.

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness |
| GET | `/health/ready` | DB readiness |

### Auth (JWT)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | — | Register customer account |
| POST | `/api/v1/auth/login` | — | Access + refresh tokens |
| POST | `/api/v1/auth/refresh` | — | Rotate refresh token |
| POST | `/api/v1/auth/logout` | — | Revoke refresh token |
| GET | `/api/v1/auth/me` | Bearer | Current user |

Password rules: min 8 characters, at least one letter and one digit.

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Secret12","first_name":"Ali","last_name":"Rahil"}'

curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Secret12"}'

curl -s http://localhost:8081/api/v1/auth/me -H "Authorization: Bearer TOKEN"
```

Refresh tokens stay on this service. Django only **verifies** access JWTs.

## Testing

```bash
make test          # unit + bdd + feature
make test-unit
make test-bdd
make test-feature
make check-ownership
```
