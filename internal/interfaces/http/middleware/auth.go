package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/response"
)

func JWTAuth(provider tokens.TokenProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "missing authorization header"))
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "invalid authorization header"))
		}

		claims, err := provider.ParseAccess(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "invalid or expired token"))
		}

		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
