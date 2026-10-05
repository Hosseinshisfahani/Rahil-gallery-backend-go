package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
)

type RegisterRequest struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func RegisterAuthRoutes(router fiber.Router, auth *service.AuthService, tokens service.JWTProvider) {
	authHandler := NewAuthHandler(auth)
	jwtAuth := JWTAuth(tokens)

	group := router.Group("/auth")
	group.Post("/register", authHandler.Register)
	group.Post("/login", authHandler.Login)
	group.Post("/refresh", authHandler.Refresh)
	group.Post("/logout", authHandler.Logout)
	group.Get("/me", jwtAuth, authHandler.Me)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	pair, err := h.auth.Register(c.Context(), service.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(OK(toAuthResponse(pair)))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	pair, err := h.auth.Login(c.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(OK(toAuthResponse(pair)))
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	pair, err := h.auth.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(OK(toAuthResponse(pair)))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid_json", "invalid request body")
	}

	if err := h.auth.Logout(c.Context(), req.RefreshToken); err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(OK(fiber.Map{"message": "logged out"}))
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

	return c.JSON(OK(profile))
}

func toAuthResponse(pair service.TokenPair) AuthResponse {
	return AuthResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        pair.TokenType,
		ExpiresIn:        int64(time.Until(pair.AccessExpiresAt).Seconds()),
		RefreshExpiresIn: int64(time.Until(pair.RefreshExpiresAt).Seconds()),
	}
}

func mapAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		return badRequest(c, "validation_error", err.Error())
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return conflict(c, "email_exists", err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return unauthorizedWithMessage(c, "invalid_credentials", err.Error())
	case errors.Is(err, service.ErrInactiveAccount):
		return forbidden(c, "inactive_account", err.Error())
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return unauthorizedWithMessage(c, "invalid_refresh_token", err.Error())
	case errors.Is(err, service.ErrPasswordNotSet):
		return forbidden(c, "password_not_set", err.Error())
	case errors.Is(err, model.ErrNotFound):
		return notFound(c, "not_found", err.Error())
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(Fail("internal_error", "something went wrong"))
	}
}

func badRequest(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(Fail(code, msg))
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Fail("unauthorized", "authentication required"))
}

func unauthorizedWithMessage(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Fail(code, msg))
}

func forbidden(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusForbidden).JSON(Fail(code, msg))
}

func conflict(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusConflict).JSON(Fail(code, msg))
}

func notFound(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(Fail(code, msg))
}

func userIDFromLocals(c *fiber.Ctx) (model.ID, error) {
	v := c.Locals("userID")
	if v == nil {
		return model.ID{}, model.ErrUnauthorized
	}
	id, ok := v.(model.ID)
	if !ok {
		return model.ID{}, model.ErrUnauthorized
	}
	return id, nil
}
