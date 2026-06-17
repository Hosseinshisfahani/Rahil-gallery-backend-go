package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	customerpg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/customer"
	identitypg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

type AdminCustomerWire struct {
	Users    identity.UserRepository
	Roles    identity.RoleRepository
	Customer customer.Repository
}

func postgresAdminCustomerWire(pool *pgxpool.Pool) AdminCustomerWire {
	return AdminCustomerWire{
		Users:    identitypg.NewUserRepository(pool),
		Roles:    identitypg.NewRoleRepository(pool),
		Customer: customerpg.NewRepository(pool),
	}
}

func RegisterAdminCustomerRoutes(router fiber.Router, tokenProvider tokens.TokenProvider, wire AdminCustomerWire, signaturesDir string) {
	svc := appcustomer.NewService(wire.Users, wire.Roles, wire.Customer, signaturesDir)
	h := handler.NewAdminCustomerHandler(svc)

	jwtAuth := middleware.JWTAuth(tokenProvider)
	staff := middleware.RequireRoles(identity.RoleAdmin, identity.RoleStaff)

	admin := router.Group("/admin/customers", jwtAuth, staff)
	admin.Get("/segments", h.ListSegments)
	admin.Get("/saved-views", h.ListSavedViews)
	admin.Post("/saved-views", h.CreateSavedView)
	admin.Get("/saved-views/:viewId", h.GetSavedView)
	admin.Patch("/saved-views/:viewId", h.UpdateSavedView)
	admin.Delete("/saved-views/:viewId", h.DeleteSavedView)

	admin.Get("/", h.List)
	admin.Post("/", h.Create)

	admin.Get("/:id", h.Get)
	admin.Patch("/:id", h.Update)
	admin.Delete("/:id", h.Delete)
	admin.Post("/:id/block", h.Block)
	admin.Post("/:id/unblock", h.Unblock)
	admin.Post("/:id/vip", h.ToggleVIP)
	admin.Post("/:id/tags", h.ToggleTag)
	admin.Post("/:id/notes", h.AddNote)
	admin.Post("/:id/signature", h.UploadSignature)
	admin.Delete("/:id/signature", h.DeleteSignature)
}
