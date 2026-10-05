package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
)

type AdminSMSHandler struct {
	jobs *repository.JobRepository
}

func NewAdminSMSHandler(jobs *repository.JobRepository) *AdminSMSHandler {
	return &AdminSMSHandler{jobs: jobs}
}

func RegisterAdminSMSRoutes(router fiber.Router, tokens service.JWTProvider, jobs *repository.JobRepository) {
	h := NewAdminSMSHandler(jobs)
	jwtAuth := JWTAuth(tokens)
	staff := RequireRoles(model.RoleAdmin, model.RoleStaff)

	admin := router.Group("/admin/sms", jwtAuth, staff)
	admin.Get("/jobs", h.ListJobs)
	admin.Patch("/jobs/:id", h.UpdateSellerNote)
}

func (h *AdminSMSHandler) ListJobs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "20"))
	result, err := h.jobs.List(c.Context(), page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Fail("internal", err.Error()))
	}
	return c.JSON(OK(fiber.Map{
		"items":      result.Items,
		"page":       result.Page,
		"perPage":    result.PerPage,
		"total":      result.Total,
		"totalPages": result.TotalPages,
	}))
}

type updateSellerNoteBody struct {
	SellerNote string `json:"sellerNote"`
}

func (h *AdminSMSHandler) UpdateSellerNote(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail("VALIDATION_ERROR", "invalid job id"))
	}
	var body updateSellerNoteBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail("INVALID_JSON", "invalid request body"))
	}
	if len([]rune(body.SellerNote)) > service.MaxSellerNoteRunes {
		return c.Status(fiber.StatusBadRequest).JSON(Fail("VALIDATION_ERROR", "sellerNote too long"))
	}
	job, err := h.jobs.UpdateSellerNote(c.Context(), id, strings.TrimSpace(body.SellerNote))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(Fail("not_found", "sms job not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(Fail("internal", err.Error()))
	}
	return c.JSON(OK(job))
}
