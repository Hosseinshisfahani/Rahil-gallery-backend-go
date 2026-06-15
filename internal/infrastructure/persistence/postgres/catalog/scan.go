package catalog

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

func scanCategory(row pgx.Row) (domain.Category, error) {
	var c domain.Category
	var parentID *uuid.UUID
	err := row.Scan(
		&c.ID, &parentID, &c.Name, &c.Slug, &c.Description,
		&c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if parentID != nil {
		id := shared.ID(*parentID)
		c.ParentID = &id
	}
	return c, err
}

func scanCollection(row pgx.Row) (domain.Collection, error) {
	var c domain.Collection
	err := row.Scan(
		&c.ID, &c.Name, &c.Slug, &c.Description,
		&c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func scanProduct(row pgx.Row) (domain.Product, error) {
	var p domain.Product
	var compareAt *float64
	var metalType *string
	var karat *int16
	var weightGrams, purityPercent *float64
	var publishedAt, deletedAt *time.Time

	err := row.Scan(
		&p.ID, &p.CategoryID, &p.SKU, &p.Name, &p.Slug,
		&p.Description, &p.ShortDescription,
		&p.JewelryType, &p.Status,
		&p.BasePrice.Amount, &compareAt, &p.BasePrice.Currency,
		&metalType, &karat, &p.GemstoneType,
		&weightGrams, &purityPercent, &p.CertificateNumber,
		&p.IsHandmade, &p.IsFeatured,
		&p.MetaTitle, &p.MetaDescription,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt, &deletedAt,
	)
	if compareAt != nil {
		p.CompareAtPrice = &shared.Money{Amount: *compareAt, Currency: p.BasePrice.Currency}
	}
	if metalType != nil {
		mt := domain.MetalType(*metalType)
		p.MetalType = &mt
	}
	p.Karat = karat
	p.WeightGrams = weightGrams
	p.PurityPercent = purityPercent
	p.PublishedAt = publishedAt
	p.DeletedAt = deletedAt
	return p, err
}

func scanVariant(row pgx.Row) (domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := row.Scan(
		&v.ID, &v.ProductID, &v.SKU, &v.Name,
		&v.SizeLabel, &v.ColorLabel, &v.PriceAdjustment,
		&v.WeightGrams, &v.Barcode,
		&v.IsDefault, &v.SortOrder, &v.IsActive,
		&v.CreatedAt, &v.UpdatedAt,
	)
	return v, err
}

func scanVariantDetail(row pgx.Row) (domain.VariantDetail, error) {
	var v domain.VariantDetail
	err := row.Scan(
		&v.ID, &v.ProductID, &v.SKU, &v.Name,
		&v.SizeLabel, &v.ColorLabel, &v.PriceAdjustment,
		&v.WeightGrams, &v.Barcode,
		&v.IsDefault, &v.SortOrder, &v.IsActive,
		&v.CreatedAt, &v.UpdatedAt,
		&v.Quantity, &v.ReservedQuantity, &v.LowStockThreshold,
	)
	return v, err
}

func scanImage(row pgx.Row) (domain.ProductImage, error) {
	var img domain.ProductImage
	var variantID *uuid.UUID
	err := row.Scan(
		&img.ID, &img.ProductID, &variantID,
		&img.URL, &img.AltText, &img.SortOrder, &img.IsPrimary, &img.CreatedAt,
	)
	if variantID != nil {
		id := shared.ID(*variantID)
		img.VariantID = &id
	}
	return img, err
}

func scanListItem(row pgx.Row) (domain.ListItem, error) {
	var item domain.ListItem
	var compareAt *float64
	var metalType *string
	var karat *int16
	var weightGrams, purityPercent *float64
	var publishedAt, deletedAt *time.Time
	var primaryImage *string

	err := row.Scan(
		&item.ID, &item.CategoryID, &item.SKU, &item.Name, &item.Slug,
		&item.Description, &item.ShortDescription,
		&item.JewelryType, &item.Status,
		&item.BasePrice.Amount, &compareAt, &item.BasePrice.Currency,
		&metalType, &karat, &item.GemstoneType,
		&weightGrams, &purityPercent, &item.CertificateNumber,
		&item.IsHandmade, &item.IsFeatured,
		&item.MetaTitle, &item.MetaDescription,
		&publishedAt, &item.CreatedAt, &item.UpdatedAt, &deletedAt,
		&item.CategoryName, &item.CategorySlug,
		&item.VariantCount, &item.PriceFrom, &item.PriceTo, &primaryImage,
	)
	if compareAt != nil {
		item.CompareAtPrice = &shared.Money{Amount: *compareAt, Currency: item.BasePrice.Currency}
	}
	if metalType != nil {
		mt := domain.MetalType(*metalType)
		item.MetalType = &mt
	}
	item.Karat = karat
	item.WeightGrams = weightGrams
	item.PurityPercent = purityPercent
	item.PublishedAt = publishedAt
	item.DeletedAt = deletedAt
	item.PrimaryImageURL = primaryImage
	return item, err
}
