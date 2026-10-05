package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool := connectDatabase(ctx, cfg.DatabaseURL)

	customers := repository.NewRepository(pool)
	jobs := repository.NewJobRepository(pool)
	smsProvider := service.NewProvider(cfg)
	tokens := service.NewJWTProvider(cfg)
	authSvc := service.NewAuthService(
		cfg,
		repository.NewUserRepository(pool),
		repository.NewRoleRepository(pool),
		repository.NewRefreshTokenRepository(pool),
		tokens,
	)
	customerSvc := service.NewCustomerService(customers, cfg.CustomerSignaturesDir)
	bulk := &service.BulkService{
		Customers: customers,
		Jobs:      jobs,
		SMS:       smsProvider,
		Sender:    cfg.KavenegarSender,
		BatchSize: cfg.SMSBulkBatchSize,
		Workers:   cfg.SMSBulkMaxConcurrency,
	}

	stopCron, err := service.StartBirthdayCron(ctx, &service.BirthdayRunner{
		Customers: customers,
		SMS:       smsProvider,
		Template:  cfg.KavenegarBirthdayTemplate,
	}, cfg.SMSBirthdayCron, cfg.SMSBirthdayTZ)
	if err != nil {
		log.Fatalf("birthday cron: %v", err)
	}
	defer stopCron()

	app := handler.NewApp(handler.AppDeps{
		Pool:      pool,
		Config:    cfg,
		Tokens:    tokens,
		Auth:      authSvc,
		Customers: customers,
		Customer:  customerSvc,
		Bulk:      bulk,
		Jobs:      jobs,
	})

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

	pool, err := repository.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	return pool
}
