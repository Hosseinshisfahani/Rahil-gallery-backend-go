package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/seed"
)

func main() {
	updatePassword := flag.Bool("update-password", false, "rotate password(s) for existing admin/staff accounts")
	allowDev := flag.Bool("allow-dev", false, "allow running when APP_ENV is not production (local testing only)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.Env != "production" && !*allowDev {
		log.Fatal("refusing to run: set APP_ENV=production or pass --allow-dev for local testing")
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("SEED_ADMIN_PASSWORD is required")
	}

	opts := seed.ProductionOptions{
		AdminEmail:     envOr("SEED_ADMIN_EMAIL", "admin@rehil.gallery"),
		AdminPassword:  adminPassword,
		AdminFirstName: envOr("SEED_ADMIN_FIRST_NAME", "Admin"),
		AdminLastName:  envOr("SEED_ADMIN_LAST_NAME", "Rehil"),
		AdminPhone:     strings.TrimSpace(os.Getenv("SEED_ADMIN_PHONE")),
		StaffEmail:     strings.TrimSpace(os.Getenv("SEED_STAFF_EMAIL")),
		StaffPassword:  os.Getenv("SEED_STAFF_PASSWORD"),
		StaffFirstName: envOr("SEED_STAFF_FIRST_NAME", "Staff"),
		StaffLastName:  envOr("SEED_STAFF_LAST_NAME", "User"),
		StaffPhone:     strings.TrimSpace(os.Getenv("SEED_STAFF_PHONE")),
		UpdatePassword: *updatePassword,
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := seed.RunProduction(ctx, pool, opts); err != nil {
		if errors.Is(err, seed.ErrProductionBootstrapped) {
			log.Println("seed-prod: skipped (admin account already exists). Use --update-password to rotate credentials.")
			os.Exit(0)
		}
		log.Fatalf("seed-prod: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
