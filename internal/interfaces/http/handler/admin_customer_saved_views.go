package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/dto"
)

func (h *AdminCustomerHandler) ListSegments(c *fiber.Ctx) error {
	staffID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}
	result, err := h.svc.ListSegments(c.Context(), staffID)
	if err != nil {
		return mapCustomerError(c, err)
	}
	return c.JSON(dto.ToSegmentsResponse(result))
}

func (h *AdminCustomerHandler) ListSavedViews(c *fiber.Ctx) error {
	staffID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}
	var viewType *domain.ListViewType
	if t := c.Query("type"); t != "" {
		vt := domain.ListViewType(t)
		viewType = &vt
	}
	views, err := h.svc.ListSavedViews(c.Context(), staffID, viewType)
	if err != nil {
		return mapCustomerError(c, err)
	}
	items := make([]dto.SavedListViewResponse, 0, len(views))
	for _, v := range views {
		items = append(items, dto.ToSavedListView(v))
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *AdminCustomerHandler) GetSavedView(c *fiber.Ctx) error {
	staffID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}
	viewID, err := parseUUIDParam(c, "viewId")
	if err != nil {
		return customerBadRequest(c, "INVALID_ID", "invalid view id")
	}
	view, err := h.svc.GetSavedView(c.Context(), staffID, viewID)
	if err != nil {
		return mapCustomerError(c, err)
	}
	return c.JSON(dto.ToSavedListView(*view))
}

func (h *AdminCustomerHandler) CreateSavedView(c *fiber.Ctx) error {
	staffID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}
	var req dto.CreateSavedViewRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	view, err := h.svc.CreateSavedView(c.Context(), staffID, savedViewInputFromRequest(req))
	if err != nil {
		return mapCustomerError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.ToSavedListView(*view))
}

func (h *AdminCustomerHandler) UpdateSavedView(c *fiber.Ctx) error {
	staffID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}
	viewID, err := parseUUIDParam(c, "viewId")
	if err != nil {
		return customerBadRequest(c, "INVALID_ID", "invalid view id")
	}
	var req dto.UpdateSavedViewRequest
	if err := c.BodyParser(&req); err != nil {
		return customerBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	view, err := h.svc.UpdateSavedView(c.Context(), staffID, viewID, savedViewInputFromUpdate(req))
	if err != nil {
		return mapCustomerError(c, err)
	}
	return c.JSON(dto.ToSavedListView(*view))
}

func (h *AdminCustomerHandler) DeleteSavedView(c *fiber.Ctx) error {
	staffID, err := userIDFromLocals(c)
	if err != nil {
		return customerUnauthorized(c)
	}
	viewID, err := parseUUIDParam(c, "viewId")
	if err != nil {
		return customerBadRequest(c, "INVALID_ID", "invalid view id")
	}
	if err := h.svc.DeleteSavedView(c.Context(), staffID, viewID); err != nil {
		return mapCustomerError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func savedViewInputFromRequest(req dto.CreateSavedViewRequest) appcustomer.SavedViewInput {
	return appcustomer.SavedViewInput{
		Name:     req.Name,
		ViewType: domain.ListViewType(req.ViewType),
		Filters:  req.Filters,
		IsShared: req.IsShared,
		Position: req.Position,
	}
}

func savedViewInputFromUpdate(req dto.UpdateSavedViewRequest) appcustomer.SavedViewInput {
	return appcustomer.SavedViewInput{
		Name:     req.Name,
		ViewType: domain.ListViewType(req.ViewType),
		Filters:  req.Filters,
		IsShared: req.IsShared,
		Position: req.Position,
	}
}

func parseUUIDParam(c *fiber.Ctx, name string) (shared.ID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return shared.ID{}, err
	}
	return shared.ID(id), nil
}
