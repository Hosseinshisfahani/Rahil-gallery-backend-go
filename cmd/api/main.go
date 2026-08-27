package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	appsms "github.com/rahil-gallery/rahil-gallery-server/internal/application/sms"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres"
	customerpg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/customer"
	infrasms "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/sms"
	httpx "github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	var pool = connectDatabase(ctx, cfg.DatabaseURL)

	smsProvider := infrasms.NewProvider(cfg)
	birthdayRunner := &appsms.BirthdayRunner{
		Customers: customerpg.NewRepository(pool),
		SMS:       smsProvider,
		Template:  cfg.KavenegarBirthdayTemplate,
	}
	stopCron, err := appsms.StartBirthdayCron(ctx, birthdayRunner, cfg.SMSBirthdayCron, cfg.SMSBirthdayTZ)
	if err != nil {
		log.Fatalf("birthday cron: %v", err)
	}
	defer stopCron()

	app := httpx.NewApp(httpx.RouterDeps{Pool: pool, Config: cfg})

	go func() {
		log.Printf("server listening on %s (env=%s)", cfg.Addr(), cfg.Env)
		if err := app.Listen(cfg.Addr()); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	if pool != nil {
		pool.Close()
	}
}

func connectDatabase(ctx context.Context, databaseURL string) *pgxpool.Pool {
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set — copy .env.example to .env, run make docker-dev, then make run")
	}

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	return pool
}
