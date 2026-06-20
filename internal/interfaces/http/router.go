package http

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	appobs "github.com/rahil-gallery/rahil-gallery-server/internal/application/observability"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/middleware"
)

func NewApp(deps RouterDeps) *fiber.App {
	var obsSvc *appobs.Service
	if deps.Pool != nil {
		obsSvc = newObservabilityService(deps.Pool, deps.Config)
	}

	app := fiber.New(fiber.Config{
		AppName:      "Rahil Gallery API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return defaultErrorHandler(c, err, obsSvc)
		},
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New())

	if deps.Config.MetricsEnabled {
		registerPrometheusMetrics(app, deps.Config)
	}

	if obsSvc != nil {
		app.Use(middleware.ObservabilityRecorder(obsSvc))
	}

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
		if obsSvc != nil {
			RegisterObservabilityRoutes(api, tokenProvider, obsSvc, deps.Config)
		}
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

func defaultErrorHandler(c *fiber.Ctx, err error, obsSvc *appobs.Service) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	if obsSvc != nil && code >= fiber.StatusInternalServerError {
		obsSvc.RecordAPIError(
			context.Background(),
			err.Error(),
			c.Path(),
			c.Method(),
			c.Get("X-Request-ID"),
			c.Get("User-Agent"),
			code,
			"",
		)
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		},
	})
}
