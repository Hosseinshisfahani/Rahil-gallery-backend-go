package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	appauth "github.com/rahil-gallery/rahil-gallery-server/internal/application/auth"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/dto"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/response"
)

type AuthHandler struct {
	auth *appauth.Service
}

func NewAuthHandler(auth *appauth.Service) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	pair, err := h.auth.Register(c.Context(), appauth.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(toAuthResponse(pair)))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	pair, err := h.auth.Login(c.Context(), appauth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(response.OK(toAuthResponse(pair)))
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	pair, err := h.auth.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(response.OK(toAuthResponse(pair)))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req dto.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	if err := h.auth.Logout(c.Context(), req.RefreshToken); err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(response.OK(fiber.Map{"message": "logged out"}))
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, err := userIDFromLocals(c)
	if err != nil {
		return unauthorized(c)
	}

	profile, err := h.auth.Me(c.Context(), userID)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(response.OK(profile))
}

func toAuthResponse(pair appauth.TokenPair) dto.AuthResponse {
	return dto.AuthResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        pair.TokenType,
		ExpiresIn:        int64(time.Until(pair.AccessExpiresAt).Seconds()),
		RefreshExpiresIn: int64(time.Until(pair.RefreshExpiresAt).Seconds()),
	}
}

func mapAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrInvalidInput):
		return badRequest(c, "validation_error", err.Error())
	case errors.Is(err, appauth.ErrEmailAlreadyExists):
		return conflict(c, "email_exists", err.Error())
	case errors.Is(err, appauth.ErrInvalidCredentials):
		return unauthorizedWithMessage(c, "invalid_credentials", err.Error())
	case errors.Is(err, appauth.ErrInactiveAccount):
		return forbidden(c, "inactive_account", err.Error())
	case errors.Is(err, appauth.ErrInvalidRefreshToken):
		return unauthorizedWithMessage(c, "invalid_refresh_token", err.Error())
	case errors.Is(err, appauth.ErrPasswordNotSet):
		return forbidden(c, "password_not_set", err.Error())
	case errors.Is(err, shared.ErrNotFound):
		return notFound(c, "not_found", err.Error())
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("internal_error", "something went wrong"))
	}
}

func badRequest(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(response.Fail(code, msg))
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication required"))
}

func unauthorizedWithMessage(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(response.Fail(code, msg))
}

func forbidden(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusForbidden).JSON(response.Fail(code, msg))
}

func conflict(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusConflict).JSON(response.Fail(code, msg))
}

func notFound(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(response.Fail(code, msg))
}

func userIDFromLocals(c *fiber.Ctx) (shared.ID, error) {
	v := c.Locals("userID")
	if v == nil {
		return shared.ID{}, shared.ErrUnauthorized
	}
	id, ok := v.(shared.ID)
	if !ok {
		return shared.ID{}, shared.ErrUnauthorized
	}
	return id, nil
}
