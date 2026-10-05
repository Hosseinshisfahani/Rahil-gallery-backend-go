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
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
)

type AdminCustomerHandler struct {
	customers *repository.Repository
	svc       *service.CustomerService
	bulkSMS   *service.BulkService
}

func NewAdminCustomerHandler(customers *repository.Repository, svc *service.CustomerService, bulkSMS *service.BulkService) *AdminCustomerHandler {
	return &AdminCustomerHandler{customers: customers, svc: svc, bulkSMS: bulkSMS}
}

func (h *AdminCustomerHandler) List(c *fiber.Ctx) error {
	filter := parseCustomerFilters(c)
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "perPage", 10)

	result, err := h.customers.List(c.Context(), filter, page, perPage)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(ToPaginatedCustomers(result))
}

func (h *AdminCustomerHandler) Create(c *fiber.Ctx) error {
	var req CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	customer, err := req.ToCustomer()
	if err != nil {
		return mapCustomerError(c, err)
	}

	customer, err = h.svc.Create(c.Context(), customer)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(ToCustomerDetail(customer))
}

func (h *AdminCustomerHandler) Get(c *fiber.Ctx) error {
	id, err := ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	customer, err := h.customers.Get(c.Context(), id)
	if err != nil {
		return mapCustomerError(c, err)
	}
	return c.JSON(ToCustomerDetail(customer))
}

func (h *AdminCustomerHandler) Update(c *fiber.Ctx) error {
	id, err := ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	var req UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	customer, err := req.ToCustomer()
	if err != nil {
		return mapCustomerError(c, err)
	}

	customer, err = h.svc.Update(c.Context(), id, customer)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(ToCustomerDetail(customer))
}

func (h *AdminCustomerHandler) Delete(c *fiber.Ctx) error {
	id, err := ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *AdminCustomerHandler) UploadSignature(c *fiber.Ctx) error {
	id, err := ParseCustomerID(c.Params("id"))
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

	data, err := io.ReadAll(io.LimitReader(f, service.MaxSignatureBytes+1))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "could not read uploaded file")
	}
	if len(data) > service.MaxSignatureBytes {
		return customerBadRequest(c, "VALIDATION_ERROR", service.ErrTooLarge.Error())
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(file.Filename))
	}

	customer, err := h.svc.UploadSignature(c.Context(), id, contentType, data)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(ToCustomerDetail(customer))
}

func (h *AdminCustomerHandler) DeleteSignature(c *fiber.Ctx) error {
	id, err := ParseCustomerID(c.Params("id"))
	if err != nil {
		return customerBadRequest(c, "VALIDATION_ERROR", "invalid customer id")
	}

	customer, err := h.svc.DeleteSignature(c.Context(), id)
	if err != nil {
		return mapCustomerError(c, err)
	}

	return c.JSON(ToCustomerDetail(customer))
}

func parseCustomerFilters(c *fiber.Ctx) model.ListFilter {
	filter := model.ListFilter{
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
	case errors.Is(err, model.ErrInvalidInput):
		return customerBadRequest(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, model.ErrNotFound):
		return customerNotFound(c, "NOT_FOUND", err.Error())
	case errors.Is(err, service.ErrPhoneAlreadyExists):
		return customerConflict(c, "PHONE_EXISTS", err.Error())
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return customerConflict(c, "EMAIL_EXISTS", err.Error())
	case errors.Is(err, model.ErrConflict):
		return customerConflict(c, "CONFLICT", err.Error())
	case errors.Is(err, service.ErrInvalidType), errors.Is(err, service.ErrTooLarge):
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

func RegisterAdminCustomerRoutes(
	router fiber.Router,
	tokens service.JWTProvider,
	customers *repository.Repository,
	svc *service.CustomerService,
	bulkSMS *service.BulkService,
) {
	h := NewAdminCustomerHandler(customers, svc, bulkSMS)
	jwtAuth := JWTAuth(tokens)
	staff := RequireRoles(model.RoleAdmin, model.RoleStaff)

	admin := router.Group("/admin/customers", jwtAuth, staff)
	admin.Get("/", h.List)
	admin.Post("/", h.Create)
	admin.Post("/sms/bulk", h.SendBulkSMS)
	admin.Get("/:id", h.Get)
	admin.Patch("/:id", h.Update)
	admin.Delete("/:id", h.Delete)
	admin.Post("/:id/signature", h.UploadSignature)
	admin.Delete("/:id/signature", h.DeleteSignature)
}

type bulkSMSBody struct {
	Message string            `json:"message"`
	Filters bulkSMSFilterBody `json:"filters"`
}

type bulkSMSFilterBody struct {
	Query            string   `json:"query"`
	CustomerID       string   `json:"customerId"`
	Email            string   `json:"email"`
	CustomerAgeRange string   `json:"customerAgeRange"`
	Gender           string   `json:"gender"`
	CustomerTypes    []string `json:"customerTypes"`
	PurchaseTypes    []string `json:"purchaseTypes"`
	FirstVisitFrom   string   `json:"firstVisitFrom"`
	FirstVisitTo     string   `json:"firstVisitTo"`
	BirthdayFrom     string   `json:"birthdayFrom"`
	BirthdayTo       string   `json:"birthdayTo"`
	MarriageFrom     string   `json:"marriageFrom"`
	MarriageTo       string   `json:"marriageTo"`
}

func (h *AdminCustomerHandler) SendBulkSMS(c *fiber.Ctx) error {
	if h.bulkSMS == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(Fail("sms_unavailable", "sms service not configured"))
	}

	var body bulkSMSBody
	if err := c.BodyParser(&body); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	filter := filterFromBulkBody(body.Filters)
	var createdBy *model.ID
	if id, err := userIDFromLocals(c); err == nil {
		createdBy = &id
	}

	result, err := h.bulkSMS.Accept(c.Context(), createdBy, body.Message, filter)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyMessage):
			return customerBadRequest(c, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, service.ErrNoTargets):
			return customerBadRequest(c, "NO_RECIPIENTS", err.Error())
		case errors.Is(err, service.ErrTooManyTargets):
			return customerBadRequest(c, "TOO_MANY", err.Error())
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(Fail("internal", err.Error()))
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(OK(result))
}

func filterFromBulkBody(b bulkSMSFilterBody) model.ListFilter {
	gender := strings.TrimSpace(b.Gender)
	if strings.EqualFold(gender, "all") {
		gender = ""
	}
	f := model.ListFilter{
		QuickSearch:      strings.TrimSpace(b.Query),
		Email:            strings.TrimSpace(b.Email),
		CustomerAgeRange: strings.TrimSpace(b.CustomerAgeRange),
		Gender:           gender,
		CustomerTypes:    b.CustomerTypes,
		PurchaseTypes:    b.PurchaseTypes,
	}
	if idStr := strings.TrimSpace(b.CustomerID); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			f.CustomerID = &id
		}
	}
	f.FirstVisitFrom, _ = model.ParseDate(strPtr(b.FirstVisitFrom))
	f.FirstVisitTo, _ = model.ParseDate(strPtr(b.FirstVisitTo))
	f.BirthdayFrom, _ = model.ParseDate(strPtr(b.BirthdayFrom))
	f.BirthdayTo, _ = model.ParseDate(strPtr(b.BirthdayTo))
	f.MarriageFrom, _ = model.ParseDate(strPtr(b.MarriageFrom))
	f.MarriageTo, _ = model.ParseDate(strPtr(b.MarriageTo))
	if f.HasAdvancedFilters() {
		f.QuickSearch = ""
	}
	return f
}

func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
