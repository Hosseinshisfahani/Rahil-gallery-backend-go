package handler

import (
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/storage/customersignature"
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

	input, err := req.ToInput()
	if err != nil {
		return mapCustomerError(c, err)
	}

	customer, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ToCustomerDetail(customer))
}

func (h *AdminCustomerHandler) Get(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	customer, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return mapCustomerError(c, err)
	}
	return c.JSON(dto.ToCustomerDetail(customer))
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

	input, err := req.ToInput()
	if err != nil {
		return mapCustomerError(c, err)
	}

	customer, err := h.svc.Update(c.Context(), id, input)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(customer))
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

func (h *AdminCustomerHandler) UploadSignature(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "signature file is required")
	}

	f, err := file.Open()
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "could not read uploaded file")
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, customersignature.MaxBytes+1))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "could not read uploaded file")
	}
	if len(data) > customersignature.MaxBytes {
		return customerBadRequest(c, "VALIDATION_ERROR", customersignature.ErrTooLarge.Error())
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(file.Filename))
	}

	customer, err := h.svc.UploadSignature(c.Context(), id, contentType, data)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(customer))
}

func (h *AdminCustomerHandler) DeleteSignature(c *fiber.Ctx) error {
	id, err := dto.ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	customer, err := h.svc.DeleteSignature(c.Context(), id)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(dto.ToCustomerDetail(customer))
}

func parseCustomerFilters(c *fiber.Ctx) domain.ListFilter {
	filter := domain.ListFilter{
		QuickSearch: strings.TrimSpace(c.Query("q")),
		Email:       strings.TrimSpace(c.Query("email")),
	}

	if idStr := strings.TrimSpace(c.Query("id")); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			filter.CustomerID = &id
		}
	}

	filter.CustomerAgeRange = strings.TrimSpace(c.Query("ageRange"))
	filter.Gender = strings.TrimSpace(c.Query("gender"))
	filter.FirstVisitFrom = queryDatePtr(c, "firstVisitFrom")
	filter.FirstVisitTo = queryDatePtr(c, "firstVisitTo")
	filter.BirthdayFrom = queryDatePtr(c, "birthdayFrom")
	filter.BirthdayTo = queryDatePtr(c, "birthdayTo")
	filter.MarriageFrom = queryDatePtr(c, "marriageFrom")
	filter.MarriageTo = queryDatePtr(c, "marriageTo")

	if customerTypes := strings.TrimSpace(c.Query("customerTypes")); customerTypes != "" {
		filter.CustomerTypes = splitCSV(customerTypes)
	}
	if purchaseTypes := strings.TrimSpace(c.Query("purchaseTypes")); purchaseTypes != "" {
		filter.PurchaseTypes = splitCSV(purchaseTypes)
	}

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

func queryFloatPtr(c *fiber.Ctx, key string) *float64 {
	v := strings.TrimSpace(c.Query(key))
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
	case errors.Is(err, customersignature.ErrInvalidType), errors.Is(err, customersignature.ErrTooLarge):
		return customerBadRequest(c, "VALIDATION_ERROR", err.Error())
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
