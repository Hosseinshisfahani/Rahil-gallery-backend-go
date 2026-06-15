package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	appcatalog "github.com/rahil-gallery/rahil-gallery-server/internal/application/catalog"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/dto"
)

type AdminProductHandler struct {
	svc *appcatalog.Service
}

func NewAdminProductHandler(svc *appcatalog.Service) *AdminProductHandler {
	return &AdminProductHandler{svc: svc}
}

func (h *AdminProductHandler) List(c *fiber.Ctx) error {
	filter := parseAdminProductFilters(c)
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "perPage", 20)

	result, err := h.svc.ListAdminProducts(c.Context(), filter, page, perPage)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToPaginatedProducts(result, false))
}

func (h *AdminProductHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return productBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	in, err := req.ToInput()
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid category id")
	}
	detail, err := h.svc.CreateProduct(c.Context(), in)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.ToProductDetail(*detail, false))
}

func (h *AdminProductHandler) Get(c *fiber.Ctx) error {
	id, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	detail, err := h.svc.GetAdminProduct(c.Context(), id)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToProductDetail(*detail, false))
}

func (h *AdminProductHandler) Update(c *fiber.Ctx) error {
	id, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	var req dto.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return productBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	in, err := req.ToInput()
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid category id")
	}
	detail, err := h.svc.UpdateProduct(c.Context(), id, in)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToProductDetail(*detail, false))
}

func (h *AdminProductHandler) Delete(c *fiber.Ctx) error {
	id, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	if err := h.svc.ArchiveProduct(c.Context(), id); err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *AdminProductHandler) CreateVariant(c *fiber.Ctx) error {
	productID, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	var req dto.CreateVariantRequest
	if err := c.BodyParser(&req); err != nil {
		return productBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	variant, err := h.svc.CreateVariant(c.Context(), productID, req.ToInput())
	if err != nil {
		return mapProductError(c, err)
	}
	product, err := h.svc.GetAdminProduct(c.Context(), productID)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.ToVariantResponse(*variant, product.BasePrice))
}

func (h *AdminProductHandler) UpdateVariant(c *fiber.Ctx) error {
	productID, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	variantID, err := dto.ParseProductID(c.Params("variantId"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid variant id")
	}
	var req dto.UpdateVariantRequest
	if err := c.BodyParser(&req); err != nil {
		return productBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	variant, err := h.svc.UpdateVariant(c.Context(), productID, variantID, req.ToInput())
	if err != nil {
		return mapProductError(c, err)
	}
	product, err := h.svc.GetAdminProduct(c.Context(), productID)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToVariantResponse(*variant, product.BasePrice))
}

func (h *AdminProductHandler) DeleteVariant(c *fiber.Ctx) error {
	productID, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	variantID, err := dto.ParseProductID(c.Params("variantId"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid variant id")
	}
	if err := h.svc.DeleteVariant(c.Context(), productID, variantID); err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *AdminProductHandler) AddImage(c *fiber.Ctx) error {
	productID, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	var req dto.CreateProductImageRequest
	if err := c.BodyParser(&req); err != nil {
		return productBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	in, err := req.ToInput()
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid variant id")
	}
	image, err := h.svc.AddImage(c.Context(), productID, in)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.ToImageResponse(*image))
}

func (h *AdminProductHandler) DeleteImage(c *fiber.Ctx) error {
	productID, err := dto.ParseProductID(c.Params("id"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid product id")
	}
	imageID, err := dto.ParseProductID(c.Params("imageId"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid image id")
	}
	if err := h.svc.DeleteImage(c.Context(), productID, imageID); err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *AdminProductHandler) AdjustInventory(c *fiber.Ctx) error {
	variantID, err := dto.ParseProductID(c.Params("variantId"))
	if err != nil {
		return productBadRequest(c, "VALIDATION_ERROR", "invalid variant id")
	}
	var req dto.AdjustInventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return productBadRequest(c, "INVALID_JSON", "invalid request body")
	}
	detail, err := h.svc.AdjustInventory(c.Context(), variantID, req.ToInput())
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToVariantResponse(*detail, shared.Money{}))
}

func parseAdminProductFilters(c *fiber.Ctx) domain.AdminListFilter {
	filter := domain.AdminListFilter{
		Query:    strings.TrimSpace(c.Query("q")),
		PriceMin: queryFloatPtr(c, "priceMin"),
		PriceMax: queryFloatPtr(c, "priceMax"),
	}
	if v := strings.TrimSpace(c.Query("status")); v != "" {
		st := domain.ProductStatus(v)
		filter.Status = &st
	}
	if v := strings.TrimSpace(c.Query("categoryId")); v != "" {
		if id, err := dto.ParseProductID(v); err == nil {
			filter.CategoryID = &id
		}
	}
	if v := strings.TrimSpace(c.Query("jewelryType")); v != "" {
		jt := domain.JewelryType(v)
		filter.JewelryType = &jt
	}
	if v := strings.TrimSpace(c.Query("metal")); v != "" {
		mt := domain.MetalType(v)
		filter.MetalType = &mt
	}
	if v := strings.TrimSpace(c.Query("gemstone")); v != "" {
		gt := domain.GemstoneType(v)
		filter.GemstoneType = &gt
	}
	if v := strings.TrimSpace(c.Query("featured")); v == "true" || v == "1" {
		b := true
		filter.Featured = &b
	} else if v == "false" || v == "0" {
		b := false
		filter.Featured = &b
	}
	return filter
}
