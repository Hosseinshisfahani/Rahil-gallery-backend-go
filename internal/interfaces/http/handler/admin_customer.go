package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/dto"
)

type AdminCustomerHandler struct {
	svc *appcustomer.Service
}

func NewAdminCustomerHandler(svc *appcustomer.Service) *AdminCustomerHandler {
	return &AdminCustomerHandler{svc: svc}
}

func (h *AdminCustomerHandler) List(c *fiber.Ctx) error {
	filter := parseCustomerFilters(c)
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "perPage", 10)

	result, err := h.svc.List(c.Context(), filter, page, perPage)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToPaginatedCustomers(result))
}

func (h *AdminCustomerHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.Create(c.Context(), adminID, req.ToInput())
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) Get(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	detail, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) Update(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	var req dto.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.Update(c.Context(), adminID, id, req.ToInput())
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) Delete(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *AdminCustomerHandler) Block(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	var req dto.BlockCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.Block(c.Context(), adminID, id, domain.BlockReason(req.Reason), req.Note)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) Unblock(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	var req dto.UnblockCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.Unblock(c.Context(), adminID, id, req.Justification)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) ToggleVIP(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.ToggleVIP(c.Context(), adminID, id)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) ToggleTag(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	var req dto.ToggleTagRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.ToggleTag(c.Context(), adminID, id, req.Tag)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func (h *AdminCustomerHandler) AddNote(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	var req dto.AddNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	adminID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}

	detail, err := h.svc.AddNote(c.Context(), adminID, id, req.Body)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(detail))
}

func parseCustomerFilters(c *fiber.Ctx) domain.ListFilter {
	filter := domain.ListFilter{
		QuickSearch:  strings.TrimSpace(c.Query("q")),
		Email:        strings.TrimSpace(c.Query("email")),
		Segment:      c.Query("segment"),
		Status:       c.Query("status"),
		HasPurchased: c.Query("hasPurchased"),
	}

	if idStr := strings.TrimSpace(c.Query("id")); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			filter.CustomerID = &id
		}
	}

	if v := c.Query("vip"); v != "" {
		b := v == "true"
		filter.VIP = &b
	}

	filter.LTVMin = queryFloatPtr(c, "ltvMin")
	filter.LTVMax = queryFloatPtr(c, "ltvMax")
	filter.OrdersMin = queryIntPtr(c, "ordersMin")
	filter.OrdersMax = queryIntPtr(c, "ordersMax")
	filter.RegisteredFrom = queryDatePtr(c, "registeredFrom")
	filter.RegisteredTo = queryDatePtr(c, "registeredTo")
	filter.LastPurchaseFrom = queryDatePtr(c, "lastPurchaseFrom")
	filter.LastPurchaseTo = queryDatePtr(c, "lastPurchaseTo")
	filter.LastActivityFrom = queryDatePtr(c, "lastActivityFrom")
	filter.LastActivityTo = queryDatePtr(c, "lastActivityTo")
	filter.CustomerAgeRange = strings.TrimSpace(c.Query("ageRange"))
	filter.Gender = strings.TrimSpace(c.Query("gender"))
	filter.FirstVisitFrom = queryDatePtr(c, "firstVisitFrom")
	filter.FirstVisitTo = queryDatePtr(c, "firstVisitTo")
	filter.BirthdayFrom = queryDatePtr(c, "birthdayFrom")
	filter.BirthdayTo = queryDatePtr(c, "birthdayTo")
	filter.MarriageFrom = queryDatePtr(c, "marriageFrom")
	filter.MarriageTo = queryDatePtr(c, "marriageTo")

	if tags := strings.TrimSpace(c.Query("tags")); tags != "" {
		filter.Tags = splitCSV(tags)
	}
	if customerTypes := strings.TrimSpace(c.Query("customerTypes")); customerTypes != "" {
		filter.CustomerTypes = splitCSV(customerTypes)
	}
	if purchaseTypes := strings.TrimSpace(c.Query("purchaseTypes")); purchaseTypes != "" {
		filter.PurchaseTypes = splitCSV(purchaseTypes)
	}

	// Quick search and advanced filters are mutually exclusive (advanced wins).
	if filter.HasAdvancedFilters() {
		filter.QuickSearch = ""
	}

	if v := strings.TrimSpace(c.Query("includeTotal")); v == "false" || v == "0" {
		filter.SkipCount = true
	}

	return filter
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	v := c.Query(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func queryIntPtr(c *fiber.Ctx, key string) *int {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

func queryFloatPtr(c *fiber.Ctx, key string) *float64 {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &n
}

func queryDatePtr(c *fiber.Ctx, key string) *time.Time {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil
	}
	return &t
}

func mapCustomerError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrInvalidInput):
		return customerBadRequest(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, shared.ErrNotFound):
		return customerNotFound(c, "NOT_FOUND", err.Error())
	case errors.Is(err, appcustomer.ErrPhoneAlreadyExists):
		return customerConflict(c, "PHONE_EXISTS", err.Error())
	case errors.Is(err, shared.ErrConflict):
		return customerConflict(c, "CONFLICT", err.Error())
	case errors.Is(err, appcustomer.ErrInvalidBlockReason), errors.Is(err, appcustomer.ErrInvalidTag),
		errors.Is(err, appcustomer.ErrInvalidSavedView), errors.Is(err, appcustomer.ErrSavedViewNameRequired):
		return customerBadRequest(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, shared.ErrForbidden):
		return customerForbidden(c)
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "something went wrong"},
		})
	}
}

func customerBadRequest(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": msg},
	})
}

func customerNotFound(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": msg},
	})
}

func customerConflict(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": msg},
	})
}

func customerUnauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": fiber.Map{"code": "UNAUTHORIZED", "message": "authentication required"},
	})
}

func customerForbidden(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error": fiber.Map{"code": "FORBIDDEN", "message": "insufficient permissions"},
	})
}
