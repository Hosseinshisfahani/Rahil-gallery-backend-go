package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	customerpg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func RegisterAdminCustomerRoutes(router fiber.Router, tokenProvider tokens.TokenProvider, pool *pgxpool.Pool, signaturesDir string) {
	repo := customerpg.NewRepository(pool)
	svc := appcustomer.NewService(repo, signaturesDir)
	h := handler.NewAdminCustomerHandler(svc)

	jwtAuth := middleware.JWTAuth(tokenProvider)
	staff := middleware.RequireRoles(identity.RoleAdmin, identity.RoleStaff)

	admin := router.Group("/admin/customers", jwtAuth, staff)
	admin.Get("/", h.List)
	admin.Post("/", h.Create)
	admin.Get("/:id", h.Get)
	admin.Patch("/:id", h.Update)
	admin.Delete("/:id", h.Delete)
	admin.Post("/:id/signature", h.UploadSignature)
	admin.Delete("/:id/signature", h.DeleteSignature)
}
