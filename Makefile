MIGRATIONS_PATH ?= migrations
MIGRATE_IMAGE ?= migrate/migrate:v4.18.1
DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:5432/rahil_gallery?sslmode=disable
# Used when Postgres runs in Docker Compose (service name: db)
MIGRATE_DATABASE_DOCKER ?= postgres://postgres:postgres@db:5432/rahil_gallery?sslmode=disable

.PHONY: test test-unit test-bdd test-feature test-integration \
	migrate-up migrate-down migrate-create migrate-up-local \
	seed seed-reset seed-small seed-products fetch-catalog-images \
	docker-up docker-up-vendor docker-vendor docker-down docker-dev docker-logs docker-migrate \
	run dev

# Local API on :8080 (requires: make docker-dev, .env with DATABASE_URL)
run dev:
	go run ./cmd/api

docker-up:
	docker compose up -d --build

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

# Dev fake data (requires: migrations applied, Postgres on localhost:5432)
# SEED_CUSTOMERS: 8 (fixtures only) … 100000 (default 10000)
# SEED_PRODUCTS:  3 (fixtures only) … 10000 (default 100)
SEED_CUSTOMERS ?= 10000
SEED_PRODUCTS  ?= 100

seed:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed --customers=$(SEED_CUSTOMERS) --products=$(SEED_PRODUCTS)

seed-reset:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed --reset --customers=$(SEED_CUSTOMERS) --products=$(SEED_PRODUCTS)

seed-small:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed --reset --customers=8 --products=3

seed-products:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed --reset --customers=8 --products=$(SEED_PRODUCTS)

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
