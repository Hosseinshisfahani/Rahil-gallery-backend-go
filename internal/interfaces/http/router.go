package http

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
)

func NewApp(deps RouterDeps) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Rahil Gallery API",
		ErrorHandler: defaultErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New())

	registerCatalogStatic(app, deps.Config.CatalogAssetsDir)
	registerCustomerSignaturesStatic(app, deps.Config.CustomerSignaturesDir)

	health := handler.NewHealthHandler(deps.Pool)
	app.Get("/health", health.Liveness)
	app.Get("/health/ready", health.Readiness)

	api := app.Group("/api/v1")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "Rahil Gallery API",
			"version": "v1",
		})
	})

	if deps.Pool != nil {
		tokenProvider := RegisterAuthRoutes(api, deps.Config, postgresAuthWire(deps.Pool))
		RegisterCatalogRoutes(api, deps.Pool)
		RegisterAdminCustomerRoutes(api, tokenProvider, postgresAdminCustomerWire(deps.Pool), deps.Config.CustomerSignaturesDir)
		RegisterAdminProductRoutes(api, tokenProvider, deps.Pool)
	}

	return app
}

func registerCatalogStatic(app *fiber.App, dir string) {
	if dir == "" {
		return
	}
	if _, err := os.Stat(dir); err != nil {
		log.Printf("catalog static assets not found at %s (run: make fetch-catalog-images)", dir)
		return
	}
	app.Static("/static/catalog", dir, fiber.Static{
		Compress:      true,
		CacheDuration: 86400,
	})
	log.Printf("serving catalog images from %s at /static/catalog", dir)
}

func registerCustomerSignaturesStatic(app *fiber.App, dir string) {
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("could not create customer signatures dir at %s: %v", dir, err)
		return
	}
	app.Static("/static/customer-signatures", dir, fiber.Static{
		Compress:      true,
		CacheDuration: 3600,
	})
	log.Printf("serving customer signatures from %s at /static/customer-signatures", dir)
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		},
	})
}
