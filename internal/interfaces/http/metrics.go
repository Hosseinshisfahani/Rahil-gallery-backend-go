package http

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
)

const (
	metricsNamespace = "rahil"
	metricsSubsystem = "api"
)

func registerPrometheusMetrics(app *fiber.App, cfg config.Config) {
	prom := fiberprometheus.NewWith(cfg.MetricsServiceName, metricsNamespace, metricsSubsystem)
	prom.SetSkipPaths([]string{
		"/metrics",
		"/health",
		"/health/ready",
	})
	prom.RegisterAt(app, "/metrics")
	app.Use(prom.Middleware)
}
