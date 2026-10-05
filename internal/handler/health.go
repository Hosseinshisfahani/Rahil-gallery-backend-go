package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return c.JSON(OK(fiber.Map{
		"status": "ok",
	}))
}

func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	if h.pool == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(Fail(
			"database_unavailable",
			"DATABASE_URL is not configured",
		))
	}

	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(Fail(
			"database_unreachable",
			"database ping failed",
		))
	}

	return c.JSON(OK(fiber.Map{
		"status":   "ready",
		"database": "up",
	}))
}
