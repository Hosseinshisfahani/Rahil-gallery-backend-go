package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                        string
	Host                       string
	Port                       int
	DatabaseURL                string
	JWTAccessSecret            string
	JWTAccessTTL               time.Duration
	JWTRefreshTTL              time.Duration
	CatalogAssetsDir           string
	CustomerSignaturesDir      string
	ObservabilityIngestKey     string
	ObservabilityRetentionDays int
	MetricsEnabled             bool
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: loading .env: %v", err)
	}

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		secret = "dev-only-change-in-production"
	}

	retentionDays, err := strconv.Atoi(getEnv("OBSERVABILITY_RETENTION_DAYS", "7"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid OBSERVABILITY_RETENTION_DAYS: %w", err)
	}

	cfg := Config{
		Env:                        getEnv("APP_ENV", "development"),
		Host:                       getEnv("APP_HOST", "0.0.0.0"),
		Port:                       port,
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		JWTAccessSecret:            secret,
		JWTAccessTTL:               accessTTL,
		JWTRefreshTTL:              refreshTTL,
		CatalogAssetsDir:           getEnv("CATALOG_ASSETS_DIR", "data/catalog-images"),
		CustomerSignaturesDir:      getEnv("CUSTOMER_SIGNATURES_DIR", "data/customer-signatures"),
		ObservabilityIngestKey:     os.Getenv("OBSERVABILITY_INGEST_KEY"),
		ObservabilityRetentionDays: retentionDays,
		MetricsEnabled:             getEnv("METRICS_ENABLED", "true") != "false",
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
