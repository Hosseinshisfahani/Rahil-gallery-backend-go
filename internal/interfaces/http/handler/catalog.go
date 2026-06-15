package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	appcatalog "github.com/rahil-gallery/rahil-gallery-server/internal/application/catalog"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/dto"
)

type CatalogHandler struct {
	svc *appcatalog.Service
}

func NewCatalogHandler(svc *appcatalog.Service) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) ListCategories(c *fiber.Ctx) error {
	items, err := h.svc.ListCategories(c.Context())
	if err != nil {
		return mapProductError(c, err)
	}
	resp := make([]dto.CategoryResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.ToCategoryResponse(item))
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (h *CatalogHandler) GetCategory(c *fiber.Ctx) error {
	item, err := h.svc.GetCategoryBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToCategoryResponse(*item))
}

func (h *CatalogHandler) ListCollections(c *fiber.Ctx) error {
	items, err := h.svc.ListCollections(c.Context())
	if err != nil {
		return mapProductError(c, err)
	}
	resp := make([]dto.CollectionResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.ToCollectionResponse(item))
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (h *CatalogHandler) GetCollection(c *fiber.Ctx) error {
	item, err := h.svc.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToCollectionResponse(*item))
}

func (h *CatalogHandler) ListProducts(c *fiber.Ctx) error {
	filter := parsePublicProductFilters(c)
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "perPage", 12)

	result, err := h.svc.ListPublishedProducts(c.Context(), filter, page, perPage)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToPaginatedProducts(result, true))
}

func (h *CatalogHandler) GetProduct(c *fiber.Ctx) error {
	detail, err := h.svc.GetPublishedProductBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToProductDetail(*detail, true))
}

func (h *CatalogHandler) ListCollectionProducts(c *fiber.Ctx) error {
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "perPage", 12)

	result, err := h.svc.ListCollectionProducts(c.Context(), c.Params("slug"), page, perPage)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(dto.ToPaginatedProducts(result, true))
}

func parsePublicProductFilters(c *fiber.Ctx) domain.PublicListFilter {
	filter := domain.PublicListFilter{
		CategorySlug:   strings.TrimSpace(c.Query("category")),
		CollectionSlug: strings.TrimSpace(c.Query("collection")),
		Query:          strings.TrimSpace(c.Query("q")),
		Sort:           strings.TrimSpace(c.Query("sort")),
		Featured:       c.Query("featured") == "true" || c.Query("featured") == "1",
		PriceMin:       queryFloatPtr(c, "priceMin"),
		PriceMax:       queryFloatPtr(c, "priceMax"),
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
	return filter
}
