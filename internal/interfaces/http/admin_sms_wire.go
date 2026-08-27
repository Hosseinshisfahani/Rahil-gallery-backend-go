package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	smspg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/sms"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func RegisterAdminSMSRoutes(router fiber.Router, tokenProvider tokens.TokenProvider, pool *pgxpool.Pool) {
	h := handler.NewAdminSMSHandler(smspg.NewJobRepository(pool))
	jwtAuth := middleware.JWTAuth(tokenProvider)
	staff := middleware.RequireRoles(identity.RoleAdmin, identity.RoleStaff)

	admin := router.Group("/admin/sms", jwtAuth, staff)
	admin.Get("/jobs", h.ListJobs)
	admin.Patch("/jobs/:id", h.UpdateSellerNote)
}
