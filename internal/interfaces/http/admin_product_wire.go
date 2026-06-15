package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	appcatalog "github.com/rahil-gallery/rahil-gallery-server/internal/application/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func RegisterAdminProductRoutes(router fiber.Router, tokenProvider tokens.TokenProvider, pool *pgxpool.Pool) {
	svc := appcatalog.NewService(postgresCatalogWire(pool))
	h := handler.NewAdminProductHandler(svc)

	jwtAuth := middleware.JWTAuth(tokenProvider)
	staff := middleware.RequireRoles(identity.RoleAdmin, identity.RoleStaff)

	admin := router.Group("/admin", jwtAuth, staff)

	products := admin.Group("/products")
	products.Get("/", h.List)
	products.Post("/", h.Create)
	products.Get("/:id", h.Get)
	products.Patch("/:id", h.Update)
	products.Delete("/:id", h.Delete)

	products.Post("/:id/variants", h.CreateVariant)
	products.Patch("/:id/variants/:variantId", h.UpdateVariant)
	products.Delete("/:id/variants/:variantId", h.DeleteVariant)

	products.Post("/:id/images", h.AddImage)
	products.Delete("/:id/images/:imageId", h.DeleteImage)

	admin.Patch("/inventory/:variantId", h.AdjustInventory)
}
