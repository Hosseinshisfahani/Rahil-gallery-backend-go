package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	appobs "github.com/rahil-gallery/rahil-gallery-server/internal/application/observability"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/observability"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/dto"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/response"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

type ObservabilityHandler struct {
	svc           *appobs.Service
	ingestKey     string
	tokenProvider tokens.TokenProvider
}

func NewObservabilityHandler(
	svc *appobs.Service,
	ingestKey string,
	tokenProvider tokens.TokenProvider,
) *ObservabilityHandler {
	return &ObservabilityHandler{svc: svc, ingestKey: ingestKey, tokenProvider: tokenProvider}
}

func (h *ObservabilityHandler) Ingest(c *fiber.Ctx) error {
	if !h.authorizeIngest(c) {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "invalid ingest credentials"))
	}

	var req dto.IngestEventRequest
	if err := c.BodyParser(&req); err != nil {
		return observabilityBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	ev, err := h.svc.Ingest(c.Context(), req.ToInput(c.Get("User-Agent")))
	if err != nil {
		return mapObservabilityError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ToObservabilityEvent(*ev))
}

func (h *ObservabilityHandler) List(c *fiber.Ctx) error {
	filter := domain.ListFilter{
		Source:  strings.TrimSpace(c.Query("source")),
		Level:   strings.TrimSpace(c.Query("level")),
		Route:   strings.TrimSpace(c.Query("route")),
		Page:    queryInt(c, "page", 1),
		PerPage: queryInt(c, "perPage", 25),
	}

	if fromStr := strings.TrimSpace(c.Query("from")); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return observabilityBadRequest(c, "VALIDATION_ERROR", "invalid from timestamp (use RFC3339)")
		}
		filter.From = &t
	}
	if toStr := strings.TrimSpace(c.Query("to")); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return observabilityBadRequest(c, "VALIDATION_ERROR", "invalid to timestamp (use RFC3339)")
		}
		filter.To = &t
	}

	events, total, err := h.svc.List(c.Context(), filter)
	if err != nil {
		return mapObservabilityError(c, err)
	}

	return c.JSON(dto.ToPaginatedObservability(events, total, filter.Page, filter.PerPage))
}

func (h *ObservabilityHandler) Summary(c *fiber.Ctx) error {
	summary, err := h.svc.Summary(c.Context())
	if err != nil {
		return mapObservabilityError(c, err)
	}
	return c.JSON(summary)
}

func (h *ObservabilityHandler) authorizeIngest(c *fiber.Ctx) bool {
	if h.ingestKey != "" && c.Get("X-Observability-Ingest-Key") == h.ingestKey {
		return true
	}

	header := c.Get("Authorization")
	if header == "" || h.tokenProvider == nil {
		return false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}

	claims, err := h.tokenProvider.ParseAccess(parts[1])
	if err != nil {
		return false
	}

	c.Locals("userID", claims.UserID)
	c.Locals("email", claims.Email)
	c.Locals("role", claims.Role)
	return true
}

func mapObservabilityError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, appobs.ErrInvalidSource),
		errors.Is(err, appobs.ErrInvalidLevel),
		errors.Is(err, appobs.ErrEmptyMessage):
		return observabilityBadRequest(c, "VALIDATION_ERROR", err.Error())
	default:
		if strings.Contains(err.Error(), "metadata must be valid JSON") {
			return observabilityBadRequest(c, "VALIDATION_ERROR", err.Error())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("internal_error", "could not process observability request"))
	}
}

func observabilityBadRequest(c *fiber.Ctx, code, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(response.Fail(code, message))
}
