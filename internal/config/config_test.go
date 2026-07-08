package config_test

import (
	"testing"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
)

func TestLoad_DefaultAppPort(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("APP_PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8081 {
		t.Fatalf("default APP_PORT = %d, want 8081", cfg.Port)
	}
}

func TestValidate_ProductionRequiresSecret(t *testing.T) {
	cfg := config.Config{
		Env:             "production",
		JWTAccessSecret: config.DefaultDevJWTSecret,
		JWTAccessTTL:    15 * time.Minute,
		JWTRefreshTTL:   24 * time.Hour,
	}

	if err := cfg.Validate(true); err == nil {
		t.Fatal("expected production config with default secret to fail validation")
	}
}

func TestValidate_ProductionAcceptsConfiguredSecret(t *testing.T) {
	cfg := config.Config{
		Env:             "production",
		JWTAccessSecret: "super-secret-production-key-32chars",
		JWTAccessTTL:    15 * time.Minute,
		JWTRefreshTTL:   24 * time.Hour,
	}

	if err := cfg.Validate(false); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}
