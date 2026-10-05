package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
	"github.com/rahil-gallery/rahil-gallery-server/internal/seed"
)

func main() {
	reset := flag.Bool("reset", false, "remove dev seed data before seeding")
	force := flag.Bool("force", false, "seed even if dev data already exists")
	customers := flag.Int("customers", seed.DefaultCustomers,
		"total customers to generate (8–100000; includes 8 fixed fixtures)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.Env == "production" {
		log.Fatal("refusing to seed: APP_ENV=production")
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	opts := seed.Options{
		Customers: *customers,
		Reset:     *reset,
		Force:     *force,
	}

	if err := seed.Run(ctx, pool, opts); err != nil {
		if errors.Is(err, seed.ErrAlreadySeeded) {
			log.Println("seed: skipped (already seeded). Use --reset or --force.")
			os.Exit(0)
		}
		log.Fatalf("seed: %v", err)
	}
}
