package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	smspg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/sms"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/response"
)

type AdminSMSHandler struct {
	jobs *smspg.JobRepository
}

func NewAdminSMSHandler(jobs *smspg.JobRepository) *AdminSMSHandler {
	return &AdminSMSHandler{jobs: jobs}
}

func (h *AdminSMSHandler) ListJobs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "20"))
	result, err := h.jobs.List(c.Context(), page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("internal", err.Error()))
	}
	return c.JSON(response.OK(fiber.Map{
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
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("VALIDATION_ERROR", "invalid job id"))
	}
	var body updateSellerNoteBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("INVALID_JSON", "invalid request body"))
	}
	if len([]rune(body.SellerNote)) > 4000 {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("VALIDATION_ERROR", "sellerNote too long"))
	}
	job, err := h.jobs.UpdateSellerNote(c.Context(), id, strings.TrimSpace(body.SellerNote))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "sms job not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("internal", err.Error()))
	}
	return c.JSON(response.OK(job))
}
