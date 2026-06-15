package dto

import (
	"github.com/google/uuid"
	appcatalog "github.com/rahil-gallery/rahil-gallery-server/internal/application/catalog"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type CreateProductRequest struct {
	CategoryID        string  `json:"categoryId"`
	SKU               string  `json:"sku"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	Description       *string `json:"description"`
	ShortDescription  *string `json:"shortDescription"`
	JewelryType       string  `json:"jewelryType"`
	Status            string  `json:"status"`
	BasePrice         float64 `json:"basePrice"`
	Currency          string  `json:"currency"`
	CompareAtPrice    *float64 `json:"compareAtPrice"`
	MetalType         *string `json:"metalType"`
	Karat             *int16  `json:"karat"`
	GemstoneType      string  `json:"gemstoneType"`
	WeightGrams       *float64 `json:"weightGrams"`
	PurityPercent     *float64 `json:"purityPercent"`
	CertificateNumber *string `json:"certificateNumber"`
	IsHandmade        bool    `json:"isHandmade"`
	IsFeatured        bool    `json:"isFeatured"`
	MetaTitle         *string `json:"metaTitle"`
	MetaDescription   *string `json:"metaDescription"`
}

type UpdateProductRequest struct {
	CategoryID        *string  `json:"categoryId"`
	SKU               *string  `json:"sku"`
	Name              *string  `json:"name"`
	Slug              *string  `json:"slug"`
	Description       *string  `json:"description"`
	ShortDescription  *string  `json:"shortDescription"`
	JewelryType       *string  `json:"jewelryType"`
	Status            *string  `json:"status"`
	BasePrice         *float64 `json:"basePrice"`
	CompareAtPrice    *float64 `json:"compareAtPrice"`
	MetalType         *string  `json:"metalType"`
	Karat             *int16   `json:"karat"`
	GemstoneType      *string  `json:"gemstoneType"`
	WeightGrams       *float64 `json:"weightGrams"`
	PurityPercent     *float64 `json:"purityPercent"`
	CertificateNumber *string  `json:"certificateNumber"`
	IsHandmade        *bool    `json:"isHandmade"`
	IsFeatured        *bool    `json:"isFeatured"`
	MetaTitle         *string  `json:"metaTitle"`
	MetaDescription   *string  `json:"metaDescription"`
}

type CreateVariantRequest struct {
	SKU               string   `json:"sku"`
	Name              string   `json:"name"`
	SizeLabel         *string  `json:"sizeLabel"`
	ColorLabel        *string  `json:"colorLabel"`
	PriceAdjustment   float64  `json:"priceAdjustment"`
	WeightGrams       *float64 `json:"weightGrams"`
	Barcode           *string  `json:"barcode"`
	IsDefault         bool     `json:"isDefault"`
	SortOrder         int      `json:"sortOrder"`
	InitialQuantity   int      `json:"initialQuantity"`
	LowStockThreshold int      `json:"lowStockThreshold"`
}

type UpdateVariantRequest struct {
	SKU             *string  `json:"sku"`
	Name            *string  `json:"name"`
	SizeLabel       *string  `json:"sizeLabel"`
	ColorLabel      *string  `json:"colorLabel"`
	PriceAdjustment *float64 `json:"priceAdjustment"`
	WeightGrams     *float64 `json:"weightGrams"`
	Barcode         *string  `json:"barcode"`
	IsDefault       *bool    `json:"isDefault"`
	SortOrder       *int     `json:"sortOrder"`
	IsActive        *bool    `json:"isActive"`
}

type CreateProductImageRequest struct {
	URL       string  `json:"url"`
	AltText   *string `json:"altText"`
	VariantID *string `json:"variantId"`
	SortOrder int     `json:"sortOrder"`
	IsPrimary bool    `json:"isPrimary"`
}

type AdjustInventoryRequest struct {
	Quantity          int  `json:"quantity"`
	LowStockThreshold *int `json:"lowStockThreshold"`
}

func ParseProductID(id string) (shared.ID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return shared.ID{}, err
	}
	return shared.ID(parsed), nil
}

func (r CreateProductRequest) ToInput() (appcatalog.CreateProductInput, error) {
	categoryID, err := ParseProductID(r.CategoryID)
	if err != nil {
		return appcatalog.CreateProductInput{}, err
	}
	status := domain.ProductStatus(r.Status)
	if status == "" {
		status = domain.ProductStatusDraft
	}
	gemstone := domain.GemstoneType(r.GemstoneType)
	if gemstone == "" {
		gemstone = domain.GemstoneTypeNone
	}
	in := appcatalog.CreateProductInput{
		CategoryID:        categoryID,
		SKU:               r.SKU,
		Name:              r.Name,
		Slug:              r.Slug,
		Description:       r.Description,
		ShortDescription:  r.ShortDescription,
		JewelryType:       domain.JewelryType(r.JewelryType),
		Status:            status,
		BasePrice:         r.BasePrice,
		Currency:          r.Currency,
		CompareAtPrice:    r.CompareAtPrice,
		Karat:             r.Karat,
		GemstoneType:      gemstone,
		WeightGrams:       r.WeightGrams,
		PurityPercent:     r.PurityPercent,
		CertificateNumber: r.CertificateNumber,
		IsHandmade:        r.IsHandmade,
		IsFeatured:        r.IsFeatured,
		MetaTitle:         r.MetaTitle,
		MetaDescription:   r.MetaDescription,
	}
	if r.MetalType != nil && *r.MetalType != "" {
		mt := domain.MetalType(*r.MetalType)
		in.MetalType = &mt
	}
	return in, nil
}

func (r UpdateProductRequest) ToInput() (appcatalog.UpdateProductInput, error) {
	in := appcatalog.UpdateProductInput{
		SKU:               r.SKU,
		Name:              r.Name,
		Slug:              r.Slug,
		Description:       r.Description,
		ShortDescription:  r.ShortDescription,
		BasePrice:         r.BasePrice,
		CompareAtPrice:    r.CompareAtPrice,
		Karat:             r.Karat,
		WeightGrams:       r.WeightGrams,
		PurityPercent:     r.PurityPercent,
		CertificateNumber: r.CertificateNumber,
		IsHandmade:        r.IsHandmade,
		IsFeatured:        r.IsFeatured,
		MetaTitle:         r.MetaTitle,
		MetaDescription:   r.MetaDescription,
	}
	if r.CategoryID != nil {
		id, err := ParseProductID(*r.CategoryID)
		if err != nil {
			return appcatalog.UpdateProductInput{}, err
		}
		in.CategoryID = &id
	}
	if r.JewelryType != nil {
		jt := domain.JewelryType(*r.JewelryType)
		in.JewelryType = &jt
	}
	if r.Status != nil {
		st := domain.ProductStatus(*r.Status)
		in.Status = &st
	}
	if r.MetalType != nil {
		mt := domain.MetalType(*r.MetalType)
		in.MetalType = &mt
	}
	if r.GemstoneType != nil {
		gt := domain.GemstoneType(*r.GemstoneType)
		in.GemstoneType = &gt
	}
	return in, nil
}

func (r CreateVariantRequest) ToInput() appcatalog.CreateVariantInput {
	return appcatalog.CreateVariantInput{
		SKU:               r.SKU,
		Name:              r.Name,
		SizeLabel:         r.SizeLabel,
		ColorLabel:        r.ColorLabel,
		PriceAdjustment:   r.PriceAdjustment,
		WeightGrams:       r.WeightGrams,
		Barcode:           r.Barcode,
		IsDefault:         r.IsDefault,
		SortOrder:         r.SortOrder,
		InitialQuantity:   r.InitialQuantity,
		LowStockThreshold: r.LowStockThreshold,
	}
}

func (r UpdateVariantRequest) ToInput() appcatalog.UpdateVariantInput {
	return appcatalog.UpdateVariantInput{
		SKU:             r.SKU,
		Name:            r.Name,
		SizeLabel:       r.SizeLabel,
		ColorLabel:      r.ColorLabel,
		PriceAdjustment: r.PriceAdjustment,
		WeightGrams:     r.WeightGrams,
		Barcode:         r.Barcode,
		IsDefault:       r.IsDefault,
		SortOrder:       r.SortOrder,
		IsActive:        r.IsActive,
	}
}

func (r CreateProductImageRequest) ToInput() (appcatalog.CreateImageInput, error) {
	in := appcatalog.CreateImageInput{
		URL:       r.URL,
		AltText:   r.AltText,
		SortOrder: r.SortOrder,
		IsPrimary: r.IsPrimary,
	}
	if r.VariantID != nil && *r.VariantID != "" {
		id, err := ParseProductID(*r.VariantID)
		if err != nil {
			return appcatalog.CreateImageInput{}, err
		}
		in.VariantID = &id
	}
	return in, nil
}

func (r AdjustInventoryRequest) ToInput() appcatalog.AdjustInventoryInput {
	return appcatalog.AdjustInventoryInput{
		Quantity:          r.Quantity,
		LowStockThreshold: r.LowStockThreshold,
	}
}
