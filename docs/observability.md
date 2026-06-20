# Observability

Application errors and client-side reports are stored in Postgres for **7 days** (configurable). Prometheus exposes API metrics at `/metrics`; Grafana dashboards can be started with Docker Compose.

## Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `OBSERVABILITY_RETENTION_DAYS` | `7` | How long events are kept |
| `OBSERVABILITY_INGEST_KEY` | _(empty)_ | Optional shared key for `X-Observability-Ingest-Key` on ingest |
| `METRICS_ENABLED` | `true` | Expose Prometheus metrics at `/metrics` |

### Endpoints

- `POST /api/v1/observability/events` — ingest (Bearer JWT **or** ingest key)
- `GET /api/v1/admin/observability/events` — staff list (filters: `source`, `level`, `route`, `page`, `perPage`)
- `GET /api/v1/admin/observability/summary` — counts for the retention window
- `GET /metrics` — Prometheus scrape target

Apply migration:

```bash
make migrate-up
```

## Prometheus + Grafana (local / VPS)

```bash
make docker-observability-up
```

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (default `admin` / `admin`, override with `GRAFANA_ADMIN_PASSWORD`)

On VPS with blocked Docker Hub, bundle images first:

```bash
bash scripts/bundle-docker-images.sh
scp /tmp/rahil-docker-images.tar.gz root@YOUR_SERVER:/root/
ssh root@YOUR_SERVER 'gunzip -c /root/rahil-docker-images.tar.gz | docker load'
```

Prometheus scrapes the API at `host.docker.internal:8080`. Ensure the API process listens on that port.

## Admin UI

Staff can browse stored errors at **Admin → Errors** (`/admin/observability`).

The Next.js admin app reports unhandled client errors when a session is active (or when `NEXT_PUBLIC_OBSERVABILITY_INGEST_KEY` is set).

## Client env

```env
# Optional — report errors without a logged-in session
# NEXT_PUBLIC_OBSERVABILITY_INGEST_KEY=
```
