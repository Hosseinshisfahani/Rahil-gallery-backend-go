package handler

import (
	"github.com/gofiber/fiber/v2"
)

func RequireRoles(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if _, ok := allowed[role]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(Fail("forbidden", "insufficient permissions"))
		}
		return c.Next()
	}
}
