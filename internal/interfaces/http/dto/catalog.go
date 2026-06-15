package dto

import (
	"math"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type CategoryResponse struct {
	ID          string  `json:"id"`
	ParentID    *string `json:"parentId,omitempty"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sortOrder"`
}

type CollectionResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
}

type ProductSummaryResponse struct {
	ID              string   `json:"id"`
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	Title           BilingualText `json:"title"`
	Category        string   `json:"category"`
	CategorySlug    string   `json:"categorySlug"`
	JewelryType     string   `json:"jewelryType"`
	BasePrice       float64  `json:"basePrice"`
	PriceFrom       float64  `json:"priceFrom"`
	PriceTo         float64  `json:"priceTo"`
	Currency        string   `json:"currency"`
	CompareAtPrice  *float64 `json:"compareAtPrice,omitempty"`
	MetalType       *string  `json:"metalType,omitempty"`
	GemstoneType    string   `json:"gemstoneType"`
	IsFeatured      bool     `json:"isFeatured"`
	IsHandmade      bool     `json:"isHandmade"`
	PrimaryImageURL *string  `json:"primaryImageUrl,omitempty"`
	Availability    string   `json:"availability"`
	VariantCount    int      `json:"variantCount"`
}

type BilingualText struct {
	En string `json:"en"`
	Fa string `json:"fa"`
}

type ProductImageResponse struct {
	ID        string  `json:"id"`
	URL       string  `json:"url"`
	AltText   *string `json:"altText,omitempty"`
	VariantID *string `json:"variantId,omitempty"`
	SortOrder int     `json:"sortOrder"`
	IsPrimary bool    `json:"isPrimary"`
}

type ProductVariantResponse struct {
	ID                string   `json:"id"`
	SKU               string   `json:"sku"`
	Name              string   `json:"name"`
	SizeLabel         *string  `json:"sizeLabel,omitempty"`
	ColorLabel        *string  `json:"colorLabel,omitempty"`
	PriceAdjustment   float64  `json:"priceAdjustment"`
	UnitPrice         float64  `json:"unitPrice"`
	WeightGrams       *float64 `json:"weightGrams,omitempty"`
	Barcode           *string  `json:"barcode,omitempty"`
	IsDefault         bool     `json:"isDefault"`
	SortOrder         int      `json:"sortOrder"`
	IsActive          bool     `json:"isActive"`
	AvailableQuantity int      `json:"availableQuantity"`
}

type ProductDetailResponse struct {
	ProductSummaryResponse
	Status            string                   `json:"status"`
	SKU               string                   `json:"sku"`
	Description       *string                  `json:"description,omitempty"`
	ShortDescription  *string                  `json:"shortDescription,omitempty"`
	Karat             *int16                   `json:"karat,omitempty"`
	WeightGrams       *float64                 `json:"weightGrams,omitempty"`
	PurityPercent     *float64                 `json:"purityPercent,omitempty"`
	CertificateNumber *string                  `json:"certificateNumber,omitempty"`
	MetaTitle         *string                  `json:"metaTitle,omitempty"`
	MetaDescription   *string                  `json:"metaDescription,omitempty"`
	PublishedAt       *string                  `json:"publishedAt,omitempty"`
	CreatedAt         string                   `json:"createdAt"`
	UpdatedAt         string                   `json:"updatedAt"`
	CategoryID        string                   `json:"categoryId"`
	Variants          []ProductVariantResponse `json:"variants"`
	Images            []ProductImageResponse   `json:"images"`
}

type PaginatedProductsResponse struct {
	Data []ProductSummaryResponse `json:"data"`
	Meta PaginationMeta           `json:"meta"`
}

func ToCategoryResponse(c domain.Category) CategoryResponse {
	resp := CategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		SortOrder:   c.SortOrder,
	}
	if c.ParentID != nil {
		s := c.ParentID.String()
		resp.ParentID = &s
	}
	return resp
}

func ToCollectionResponse(c domain.Collection) CollectionResponse {
	return CollectionResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
	}
}

func ToPaginatedProducts(result domain.ListResult, public bool) PaginatedProductsResponse {
	items := make([]ProductSummaryResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, ToProductSummary(item, public))
	}
	totalPages := int(math.Ceil(float64(result.Total) / float64(result.PerPage)))
	return PaginatedProductsResponse{
		Data: items,
		Meta: PaginationMeta{
			Page:       result.Page,
			PerPage:    result.PerPage,
			Total:      &result.Total,
			TotalPages: &totalPages,
		},
	}
}

func ToProductSummary(item domain.ListItem, public bool) ProductSummaryResponse {
	resp := ProductSummaryResponse{
		ID:              item.ID.String(),
		Slug:            item.Slug,
		Name:            item.Name,
		Title:           bilingualTitle(item.Name),
		Category:        item.CategoryName,
		CategorySlug:    item.CategorySlug,
		JewelryType:     string(item.JewelryType),
		BasePrice:       item.BasePrice.Amount,
		PriceFrom:       item.PriceFrom,
		PriceTo:         item.PriceTo,
		Currency:        item.BasePrice.Currency,
		GemstoneType:    string(item.GemstoneType),
		IsFeatured:      item.IsFeatured,
		IsHandmade:      item.IsHandmade,
		PrimaryImageURL: item.PrimaryImageURL,
		VariantCount:    item.VariantCount,
	}
	if item.CompareAtPrice != nil {
		resp.CompareAtPrice = &item.CompareAtPrice.Amount
	}
	if item.MetalType != nil {
		s := string(*item.MetalType)
		resp.MetalType = &s
	}
	if public {
		resp.Availability = string(item.Availability)
	} else if item.Status == domain.ProductStatusPublished {
		resp.Availability = string(domain.AvailabilityInStock)
	}
	return resp
}

func ToProductDetail(detail domain.ProductDetail, public bool) ProductDetailResponse {
	summary := ToProductSummary(domain.ListItem{
		Product:         detail.Product,
		CategoryName:    detail.Category.Name,
		CategorySlug:    detail.Category.Slug,
		VariantCount:    len(detail.Variants),
		PriceFrom:       0,
		PriceTo:         0,
		PrimaryImageURL: primaryImageURL(detail.Images),
		Availability:    domain.DeriveAvailability(detail.Variants),
	}, public)
	from, to := domain.PriceRange(detail.BasePrice, detail.Variants)
	summary.PriceFrom = from
	summary.PriceTo = to

	resp := ProductDetailResponse{
		ProductSummaryResponse: summary,
		Status:                 string(detail.Status),
		SKU:                    detail.SKU,
		Description:            detail.Description,
		ShortDescription:       detail.ShortDescription,
		Karat:                  detail.Karat,
		WeightGrams:            detail.WeightGrams,
		PurityPercent:          detail.PurityPercent,
		CertificateNumber:      detail.CertificateNumber,
		MetaTitle:              detail.MetaTitle,
		MetaDescription:        detail.MetaDescription,
		CreatedAt:              detail.CreatedAt.UTC().Format(timeRFC3339),
		UpdatedAt:              detail.UpdatedAt.UTC().Format(timeRFC3339),
		CategoryID:             detail.CategoryID.String(),
	}
	if detail.PublishedAt != nil {
		s := detail.PublishedAt.UTC().Format(timeRFC3339)
		resp.PublishedAt = &s
	}
	if detail.MetalType != nil {
		s := string(*detail.MetalType)
		resp.MetalType = &s
	}
	resp.Variants = make([]ProductVariantResponse, 0, len(detail.Variants))
	for _, v := range detail.Variants {
		if public && !v.IsActive {
			continue
		}
		resp.Variants = append(resp.Variants, ToVariantResponse(v, detail.BasePrice))
	}
	resp.Images = make([]ProductImageResponse, 0, len(detail.Images))
	for _, img := range detail.Images {
		resp.Images = append(resp.Images, ToImageResponse(img))
	}
	return resp
}

func ToVariantResponse(v domain.VariantDetail, base shared.Money) ProductVariantResponse {
	unit := v.UnitPrice(base)
	return ProductVariantResponse{
		ID:                v.ID.String(),
		SKU:               v.SKU,
		Name:              v.Name,
		SizeLabel:         v.SizeLabel,
		ColorLabel:        v.ColorLabel,
		PriceAdjustment:   v.PriceAdjustment,
		UnitPrice:         unit.Amount,
		WeightGrams:       v.WeightGrams,
		Barcode:           v.Barcode,
		IsDefault:         v.IsDefault,
		SortOrder:         v.SortOrder,
		IsActive:          v.IsActive,
		AvailableQuantity: v.Available(),
	}
}

func ToImageResponse(img domain.ProductImage) ProductImageResponse {
	resp := ProductImageResponse{
		ID:        img.ID.String(),
		URL:       img.URL,
		AltText:   img.AltText,
		SortOrder: img.SortOrder,
		IsPrimary: img.IsPrimary,
	}
	if img.VariantID != nil {
		s := img.VariantID.String()
		resp.VariantID = &s
	}
	return resp
}

func bilingualTitle(name string) BilingualText {
	return BilingualText{En: name, Fa: name}
}

func primaryImageURL(images []domain.ProductImage) *string {
	for _, img := range images {
		if img.IsPrimary {
			return &img.URL
		}
	}
	if len(images) > 0 {
		return &images[0].URL
	}
	return nil
}

const timeRFC3339 = "2006-01-02T15:04:05Z"
