package http

import (
	"github.com/gofiber/fiber/v2"
	appauth "github.com/rahil-gallery/rahil-gallery-server/internal/application/auth"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/security"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
)

type AuthWire struct {
	Users         identity.UserRepository
	Roles         identity.RoleRepository
	RefreshTokens identity.RefreshTokenRepository
}

func BuildAuthService(cfg config.Config, wire AuthWire) *appauth.Service {
	tokenProvider := security.NewJWTProvider(cfg)
	passwordHasher := security.NewBcryptHasher()
	return appauth.NewService(cfg, wire.Users, wire.Roles, wire.RefreshTokens, tokenProvider, passwordHasher)
}

func RegisterAuthRoutes(router fiber.Router, cfg config.Config, wire AuthWire) tokens.TokenProvider {
	svc := BuildAuthService(cfg, wire)
	authHandler := handler.NewAuthHandler(svc)
	tokenProvider := security.NewJWTProvider(cfg)
	jwtAuth := middleware.JWTAuth(tokenProvider)

	auth := router.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", jwtAuth, authHandler.Me)

	return tokenProvider
}
