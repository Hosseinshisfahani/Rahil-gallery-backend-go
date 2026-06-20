package main

import (
	"context"
	"log"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/seed"
)

// Backfill CRM demographics (age group, gender, type, categories) on existing customers.
// Safe to run in production after migrations.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := seed.BackfillCRMProfiles(ctx, pool); err != nil {
		log.Fatalf("backfill-crm: %v", err)
	}
	log.Println("backfill-crm: done")
}
