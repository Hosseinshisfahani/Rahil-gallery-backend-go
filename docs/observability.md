# Observability

Application errors and client-side reports are stored in Postgres for **7 days** (configurable). Prometheus exposes API metrics at `/metrics`; Grafana dashboards are started automatically on deploy.

## Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `OBSERVABILITY_RETENTION_DAYS` | `7` | How long events are kept |
| `OBSERVABILITY_INGEST_KEY` | _(empty)_ | Optional shared key for `X-Observability-Ingest-Key` on ingest |
| `METRICS_ENABLED` | `true` | Expose Prometheus metrics at `/metrics` |
| `METRICS_SERVICE_NAME` | `rahil_api` | Prometheus `service` label on HTTP metrics |
| `OBSERVABILITY_ENABLED` | `true` | Start Prometheus + Grafana during deploy |

### Endpoints

- `POST /api/v1/observability/events` — ingest (Bearer JWT **or** ingest key)
- `GET /metrics` — Prometheus scrape target

## CI/CD (automatic on push to `master`)

The server deploy workflow syncs observability config and `deploy-remote.sh`:

1. Runs DB migrations (includes `observability_events`)
2. Starts the API (`METRICS_ENABLED=true`)
3. Starts **Prometheus** (`:9090`) and **Grafana** (`:3001`)
4. Reloads Prometheus config after each deploy

### One-time VPS setup (blocked Docker Hub)

Load base images once (Postgres, migrate, Prometheus, Grafana):

```bash
bash scripts/bundle-docker-images.sh
scp /tmp/rahil-docker-images.tar.gz root@YOUR_SERVER:/root/
ssh root@YOUR_SERVER 'gunzip -c /root/rahil-docker-images.tar.gz | docker load'
```

Set in server `.env`:

```env
JWT_ACCESS_SECRET=...
GRAFANA_ADMIN_PASSWORD=change-me
OBSERVABILITY_ENABLED=true
```

## Local development

```bash
make docker-dev                    # Postgres
make run                           # API on :8080
make docker-observability-up       # Prometheus + Grafana
```

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (`admin` / `admin`, or `GRAFANA_ADMIN_PASSWORD`)

Prometheus scrapes the API at `127.0.0.1:8080` (host network). Grafana reads Postgres via the `db` service on the compose network.

## Dashboards

| Dashboard | Source | Content |
|-----------|--------|---------|
| **Rahil Gallery API** | Prometheus | Request rate, latency, in-flight |
| **Rahil Gallery Errors** | Postgres | Application + client errors (7-day window) |

## Client error reporting

The Next.js admin app reports unhandled client errors when a session is active (or when `NEXT_PUBLIC_OBSERVABILITY_INGEST_KEY` is set). View errors in Grafana, not in the admin UI.

```env
# Optional — report errors without a logged-in session
# NEXT_PUBLIC_OBSERVABILITY_INGEST_KEY=
```

## Disable observability stack on deploy

```env
OBSERVABILITY_ENABLED=false
```

The API still records errors and exposes `/metrics`; only Prometheus/Grafana containers are skipped.
