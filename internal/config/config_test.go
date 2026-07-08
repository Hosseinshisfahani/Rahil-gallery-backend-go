package config_test

import (
	"testing"

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
