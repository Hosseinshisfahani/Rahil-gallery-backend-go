package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/response"
)

func RequireRoles(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if _, ok := allowed[role]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "insufficient permissions"))
		}
		return c.Next()
	}
}
