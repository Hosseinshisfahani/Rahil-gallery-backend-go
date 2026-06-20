package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	appobs "github.com/rahil-gallery/rahil-gallery-server/internal/application/observability"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	observabilitypg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/observability"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func newObservabilityService(pool *pgxpool.Pool, cfg config.Config) *appobs.Service {
	repo := observabilitypg.NewRepository(pool)
	svc := appobs.NewService(repo, cfg.ObservabilityRetentionDays)
	svc.StartBackgroundJobs(context.Background())
	return svc
}

func RegisterObservabilityRoutes(
	router fiber.Router,
	tokenProvider tokens.TokenProvider,
	svc *appobs.Service,
	cfg config.Config,
) {
	h := handler.NewObservabilityHandler(svc, cfg.ObservabilityIngestKey, tokenProvider)

	router.Post("/observability/events", h.Ingest)

	jwtAuth := middleware.JWTAuth(tokenProvider)
	staff := middleware.RequireRoles(identity.RoleAdmin, identity.RoleStaff)

	admin := router.Group("/admin/observability", jwtAuth, staff)
	admin.Get("/events", h.List)
	admin.Get("/summary", h.Summary)
}
