package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	appsms "github.com/rahil-gallery/rahil-gallery-server/internal/application/sms"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/response"
)

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
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("sms_unavailable", "sms service not configured"))
	}

	var body bulkSMSBody
	if err := c.BodyParser(&body); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}

	filter := filterFromBulkBody(body.Filters)
	var createdBy *shared.ID
	if id, err := userIDFromLocals(c); err == nil {
		createdBy = &id
	}

	result, err := h.bulkSMS.Accept(c.Context(), createdBy, body.Message, filter)
	if err != nil {
		switch {
		case errors.Is(err, appsms.ErrEmptyMessage):
			return customerBadRequest(c, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, appsms.ErrNoTargets):
			return customerBadRequest(c, "NO_RECIPIENTS", err.Error())
		case errors.Is(err, appsms.ErrTooManyTargets):
			return customerBadRequest(c, "TOO_MANY", err.Error())
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("internal", err.Error()))
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(response.OK(result))
}

func filterFromBulkBody(b bulkSMSFilterBody) domain.ListFilter {
	gender := strings.TrimSpace(b.Gender)
	if strings.EqualFold(gender, "all") {
		gender = ""
	}
	f := domain.ListFilter{
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
	f.FirstVisitFrom, _ = domain.ParseDate(strPtr(b.FirstVisitFrom))
	f.FirstVisitTo, _ = domain.ParseDate(strPtr(b.FirstVisitTo))
	f.BirthdayFrom, _ = domain.ParseDate(strPtr(b.BirthdayFrom))
	f.BirthdayTo, _ = domain.ParseDate(strPtr(b.BirthdayTo))
	f.MarriageFrom, _ = domain.ParseDate(strPtr(b.MarriageFrom))
	f.MarriageTo, _ = domain.ParseDate(strPtr(b.MarriageTo))
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
