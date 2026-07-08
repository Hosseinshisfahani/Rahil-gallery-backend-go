//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres"
	httpx "github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http"
)

func TestAuthIntegration_RegisterAndLogin(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	cfg := config.Config{
		Env:             "test",
		JWTAccessSecret: "integration-test-secret-key-32chars",
		JWTAccessTTL:    15 * time.Minute,
		JWTRefreshTTL:   24 * time.Hour,
	}

	app := httpx.NewApp(httpx.RouterDeps{Pool: pool, Config: cfg})

	email := "integration-" + time.Now().Format("150405") + "@example.com"
	regBody := map[string]string{
		"email": email, "password": "Secret12",
		"first_name": "Int", "last_name": "Test",
	}
	body, _ := json.Marshal(regBody)
	reg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	reg.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(reg, -1)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("register status %d", regResp.StatusCode)
	}

	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": "Secret12"})
	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	login.Header.Set("Content-Type", "application/json")
	loginResp, _ := app.Test(login, -1)
	if loginResp.StatusCode == http.StatusNotFound {
		t.Fatal("POST /api/v1/auth/login must be registered (got 404)")
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status %d", loginResp.StatusCode)
	}
}
