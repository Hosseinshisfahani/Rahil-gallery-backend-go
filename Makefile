MIGRATIONS_PATH ?= migrations
MIGRATE_IMAGE ?= migrate/migrate:v4.18.1
POSTGRES_PORT ?= 5433
APP_PORT ?= 8081
DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:$(POSTGRES_PORT)/rahil_gallery?sslmode=disable
export POSTGRES_PORT APP_PORT
# Used when Postgres runs in Docker Compose (service name: db)
MIGRATE_DATABASE_DOCKER ?= postgres://postgres:postgres@db:5432/rahil_gallery?sslmode=disable
# VPS: proxy.golang.org is often blocked — use goproxy.io or vendor/offline binary instead.
GOPROXY ?= https://goproxy.io,https://goproxy.cn,direct
export GOPROXY

.PHONY: test test-unit test-bdd test-feature test-integration \
	migrate-up migrate-down migrate-create migrate-up-local \
	seed seed-reset seed-small seed-products seed-prod seed-prod-password fetch-catalog-images \
	docker-up docker-up-vendor docker-vendor docker-down docker-dev docker-prod-up docker-prod-migrate docker-prod-api docker-observability-up docker-logs docker-migrate \
	run dev dev-check run-docker stop-api run-vendor build-linux run-binary

# Local API on :8081 by default (requires: make docker-dev)
dev-check:
	bash scripts/check-dev-db.sh

dev: dev-check
	bash scripts/run-api-dev.sh run

run:
	bash scripts/run-api-dev.sh run

run-docker:
	bash scripts/run-api-dev.sh start

stop-api:
	bash scripts/run-api-dev.sh stop

run-vendor:
	go run -mod=vendor ./cmd/api

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/rahil-api ./cmd/api

# Linux CLI tools shipped to VPS by CI (no Go compiler required on server).
build-prod-tools:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/seed-prod ./cmd/seed-prod

PROD_SEED_CMD = $(if $(wildcard bin/seed-prod),./bin/seed-prod,go run ./cmd/seed-prod)

run-binary:
	./bin/rahil-api

# Dev machine only — pulls golang/alpine from Docker Hub (fails on blocked VPS).
docker-up:
	docker compose up -d --build

# VPS / production — no build; uses pre-loaded image from CI/CD (rahil-gallery-api:latest).
docker-prod-up:
	eval "$$(bash scripts/export-compose-env.sh)" && docker compose -f docker-compose.prod.yml up -d

docker-prod-migrate:
	eval "$$(bash scripts/export-compose-env.sh)" && docker compose -f docker-compose.prod.yml up migrate --abort-on-container-exit

docker-prod-api:
	eval "$$(bash scripts/export-compose-env.sh)" && docker compose -f docker-compose.prod.yml up -d --force-recreate api

docker-observability-up:
	eval "$$(bash scripts/export-compose-env.sh)" && docker compose -f docker-compose.prod.yml -f docker-compose.observability.yml up -d prometheus grafana

docker-vendor:
	go mod vendor

docker-up-vendor: docker-vendor
	docker compose -f docker-compose.yml -f docker-compose.vendor.yml up -d --build

docker-down:
	docker compose down

docker-dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d db migrate

docker-logs:
	docker compose logs -f api

# Alias for migrate-up (requires: docker compose up -d db)
docker-migrate: migrate-up

test:
	go test ./... -count=1

test-unit:
	go test ./internal/... -count=1

test-bdd:
	go test ./tests/bdd/... -count=1

test-feature:
	go test ./tests/feature/... -count=1

test-integration:
	go test -tags=integration ./tests/integration/... -count=1

# Apply migrations (Postgres must be running — e.g. make docker-dev)
migrate-up:
	docker compose run --rm --no-deps migrate \
		-path=/migrations \
		-database "$(MIGRATE_DATABASE_DOCKER)" \
		up

migrate-down:
	docker compose run --rm --no-deps migrate \
		-path=/migrations \
		-database "$(MIGRATE_DATABASE_DOCKER)" \
		down 1

# Create migration files (no DB required). Usage: make migrate-create NAME=add_products
migrate-create:
ifndef NAME
	@read -p "Migration name: " name; \
	docker run --rm \
		-v "$(CURDIR)/$(MIGRATIONS_PATH):/migrations" \
		$(MIGRATE_IMAGE) \
		create -ext sql -dir /migrations -seq $$name
else
	docker run --rm \
		-v "$(CURDIR)/$(MIGRATIONS_PATH):/migrations" \
		$(MIGRATE_IMAGE) \
		create -ext sql -dir /migrations -seq $(NAME)
endif

# Dev fake data (requires: make docker-dev, migrations applied)
# Uses scripts/run-seed.sh → Compose network db:5432 (host :5432 publish can hang on some Docker setups).
# SEED_CUSTOMERS: 8 (fixtures only) … 100000 (default 10000)
# SEED_PRODUCTS:  3 (fixtures only) … 10000 (default 100)
SEED_CUSTOMERS ?= 10000
SEED_PRODUCTS  ?= 100

seed:
	bash scripts/run-seed.sh --customers=$(SEED_CUSTOMERS) --products=$(SEED_PRODUCTS)

seed-reset:
	bash scripts/run-seed.sh --reset --customers=$(SEED_CUSTOMERS) --products=$(SEED_PRODUCTS)

seed-small:
	bash scripts/run-seed.sh --reset --customers=8 --products=3

seed-products:
	bash scripts/run-seed.sh --reset --customers=8 --products=$(SEED_PRODUCTS)

# Production bootstrap — admin account only (no demo customers/catalog).
# Requires APP_ENV=production (or --allow-dev for local smoke tests).
seed-prod:
	DATABASE_URL="$(DATABASE_URL)" $(PROD_SEED_CMD)

seed-prod-password:
	DATABASE_URL="$(DATABASE_URL)" $(PROD_SEED_CMD) --update-password

fetch-catalog-images:
	bash scripts/fetch-catalog-images.sh

# Migrations against Postgres on host port 5432 (Linux: --network host)
migrate-up-local:
	docker run --rm --network host \
		-v "$(CURDIR)/$(MIGRATIONS_PATH):/migrations" \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$(DATABASE_URL)" \
		up
