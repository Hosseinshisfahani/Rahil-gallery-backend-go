package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	appsms "github.com/rahil-gallery/rahil-gallery-server/internal/application/sms"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	customerpg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/customer"
	smspg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/sms"
	infrasms "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/sms"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func RegisterAdminCustomerRoutes(router fiber.Router, tokenProvider tokens.TokenProvider, pool *pgxpool.Pool, cfg config.Config) {
	repo := customerpg.NewRepository(pool)
	svc := appcustomer.NewService(repo, cfg.CustomerSignaturesDir)
	bulkSMS := &appsms.BulkService{
		Customers: repo,
		Jobs:      smspg.NewJobRepository(pool),
		SMS:       infrasms.NewProvider(cfg),
		Sender:    cfg.KavenegarSender,
		BatchSize: cfg.SMSBulkBatchSize,
		Workers:   cfg.SMSBulkMaxConcurrency,
	}
	h := handler.NewAdminCustomerHandler(svc, bulkSMS)

	jwtAuth := middleware.JWTAuth(tokenProvider)
	staff := middleware.RequireRoles(identity.RoleAdmin, identity.RoleStaff)

	admin := router.Group("/admin/customers", jwtAuth, staff)
	admin.Get("/", h.List)
	admin.Post("/", h.Create)
	admin.Post("/sms/bulk", h.SendBulkSMS)
	admin.Get("/:id", h.Get)
	admin.Patch("/:id", h.Update)
	admin.Delete("/:id", h.Delete)
	admin.Post("/:id/signature", h.UploadSignature)
	admin.Delete("/:id/signature", h.DeleteSignature)
}
