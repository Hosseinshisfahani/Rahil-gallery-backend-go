package testauth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/test/fakeidentity"
	httpx "github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http"
)

func TestConfig() config.Config {
	return config.Config{
		Env:             "test",
		Host:            "127.0.0.1",
		Port:            0,
		JWTAccessSecret: "test-secret-key-at-least-32-chars-long",
		JWTAccessTTL:    15 * time.Minute,
		JWTRefreshTTL:   24 * time.Hour,
	}
}

func NewApp() (*fiber.App, *fakeidentity.Store) {
	store := fakeidentity.NewStore()
	cfg := TestConfig()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	api := app.Group("/api/v1")
	httpx.RegisterAuthRoutes(api, cfg, WireFromStore(store))
	return app, store
}

func WireFromStore(store *fakeidentity.Store) httpx.AuthWire {
	return httpx.AuthWire{
		Users:         fakeidentity.UserRepo{S: store},
		Roles:         fakeidentity.RoleRepo{S: store},
		RefreshTokens: fakeidentity.RefreshRepo{S: store},
	}
}

var (
	_ identity.UserRepository         = fakeidentity.UserRepo{}
	_ identity.RoleRepository         = fakeidentity.RoleRepo{}
	_ identity.RefreshTokenRepository = fakeidentity.RefreshRepo{}
)
