package middleware

import (
	"github.com/gofiber/fiber/v2"
	appobs "github.com/rahil-gallery/rahil-gallery-server/internal/application/observability"
)

func ObservabilityRecorder(svc *appobs.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if svc == nil {
			return err
		}

		status := c.Response().StatusCode()
		if err == nil && status >= fiber.StatusInternalServerError {
			message := "request failed"
			if err != nil {
				message = err.Error()
			}
			svc.RecordAPIError(
				c.Context(),
				message,
				c.Path(),
				c.Method(),
				c.Get("X-Request-ID"),
				c.Get("User-Agent"),
				status,
				"",
			)
		}

		return err
	}
}
