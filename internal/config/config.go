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
	Env               string
	Host              string
	Port              int
	DatabaseURL       string
	JWTAccessSecret   string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	CatalogAssetsDir  string
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

	cfg := Config{
		Env:              getEnv("APP_ENV", "development"),
		Host:             getEnv("APP_HOST", "0.0.0.0"),
		Port:             port,
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTAccessSecret:  secret,
		JWTAccessTTL:     accessTTL,
		JWTRefreshTTL:    refreshTTL,
		CatalogAssetsDir: getEnv("CATALOG_ASSETS_DIR", "data/catalog-images"),
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
