package catalog

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

var (
	ErrSKURequired       = errors.New("sku is required")
	ErrNameRequired      = errors.New("name is required")
	ErrCategoryRequired  = errors.New("category is required")
	ErrInvalidStatus     = errors.New("invalid product status")
	ErrInvalidJewelry    = errors.New("invalid jewelry type")
	ErrInvalidMetal      = errors.New("invalid metal type")
	ErrInvalidGemstone   = errors.New("invalid gemstone type")
	ErrVariantRequired   = errors.New("variant name is required")
	ErrImageURLRequired  = errors.New("image url is required")
	ErrNegativePrice     = errors.New("price must be non-negative")
	ErrNegativeInventory = errors.New("inventory quantity must be non-negative")
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

type CreateProductInput struct {
	CategoryID        shared.ID
	SKU               string
	Name              string
	Slug              string
	Description       *string
	ShortDescription  *string
	JewelryType       domain.JewelryType
	Status            domain.ProductStatus
	BasePrice         float64
	Currency          string
	CompareAtPrice    *float64
	MetalType         *domain.MetalType
	Karat             *int16
	GemstoneType      domain.GemstoneType
	WeightGrams       *float64
	PurityPercent     *float64
	CertificateNumber *string
	IsHandmade        bool
	IsFeatured        bool
	MetaTitle         *string
	MetaDescription   *string
}

type UpdateProductInput struct {
	CategoryID        *shared.ID
	SKU               *string
	Name              *string
	Slug              *string
	Description       *string
	ShortDescription  *string
	JewelryType       *domain.JewelryType
	Status            *domain.ProductStatus
	BasePrice         *float64
	CompareAtPrice    *float64
	MetalType         *domain.MetalType
	Karat             *int16
	GemstoneType      *domain.GemstoneType
	WeightGrams       *float64
	PurityPercent     *float64
	CertificateNumber *string
	IsHandmade        *bool
	IsFeatured        *bool
	MetaTitle         *string
	MetaDescription   *string
}

type CreateVariantInput struct {
	SKU               string
	Name              string
	SizeLabel         *string
	ColorLabel        *string
	PriceAdjustment   float64
	WeightGrams       *float64
	Barcode           *string
	IsDefault         bool
	SortOrder         int
	InitialQuantity   int
	LowStockThreshold int
}

type UpdateVariantInput struct {
	SKU             *string
	Name            *string
	SizeLabel       *string
	ColorLabel      *string
	PriceAdjustment *float64
	WeightGrams     *float64
	Barcode         *string
	IsDefault       *bool
	SortOrder       *int
	IsActive        *bool
}

type CreateImageInput struct {
	URL       string
	AltText   *string
	VariantID *shared.ID
	SortOrder int
	IsPrimary bool
}

type AdjustInventoryInput struct {
	Quantity          int
	LowStockThreshold *int
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.ListActiveCategories(ctx)
}

func (s *Service) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return s.repo.FindCategoryBySlug(ctx, slug)
}

func (s *Service) ListCollections(ctx context.Context) ([]domain.Collection, error) {
	return s.repo.ListActiveCollections(ctx)
}

func (s *Service) GetCollectionBySlug(ctx context.Context, slug string) (*domain.Collection, error) {
	return s.repo.FindCollectionBySlug(ctx, slug)
}

func (s *Service) ListPublishedProducts(ctx context.Context, filter domain.PublicListFilter, page, perPage int) (domain.ListResult, error) {
	return s.repo.ListPublished(ctx, filter, page, perPage)
}

func (s *Service) ListCollectionProducts(ctx context.Context, slug string, page, perPage int) (domain.ListResult, error) {
	col, err := s.repo.FindCollectionBySlug(ctx, slug)
	if err != nil {
		return domain.ListResult{}, err
	}
	return s.repo.ListPublishedByCollection(ctx, col.ID, page, perPage)
}

func (s *Service) GetPublishedProductBySlug(ctx context.Context, slug string) (*domain.ProductDetail, error) {
	product, err := s.repo.FindPublishedProductBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return s.buildDetail(ctx, product)
}

func (s *Service) ListAdminProducts(ctx context.Context, filter domain.AdminListFilter, page, perPage int) (domain.ListResult, error) {
	return s.repo.ListAdmin(ctx, filter, page, perPage)
}

func (s *Service) GetAdminProduct(ctx context.Context, id shared.ID) (*domain.ProductDetail, error) {
	product, err := s.repo.FindProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.buildDetail(ctx, product)
}

func (s *Service) CreateProduct(ctx context.Context, in CreateProductInput) (*domain.ProductDetail, error) {
	if err := validateCreateProduct(in); err != nil {
		return nil, err
	}
	slug := normalizeSlug(in.Slug, in.Name)
	if exists, err := s.repo.SlugExists(ctx, slug, nil); err != nil {
		return nil, err
	} else if exists {
		return nil, shared.ErrConflict
	}
	if exists, err := s.repo.SKUExists(ctx, strings.TrimSpace(in.SKU), nil); err != nil {
		return nil, err
	} else if exists {
		return nil, shared.ErrConflict
	}
	if _, err := s.repo.FindCategoryByID(ctx, in.CategoryID); err != nil {
		return nil, err
	}

	currency := strings.TrimSpace(in.Currency)
	if currency == "" {
		currency = "IRR"
	}

	product := &domain.Product{
		ID:                shared.ID(uuid.New()),
		CategoryID:        in.CategoryID,
		SKU:               strings.TrimSpace(in.SKU),
		Name:              strings.TrimSpace(in.Name),
		Slug:              slug,
		Description:       in.Description,
		ShortDescription:  in.ShortDescription,
		JewelryType:       in.JewelryType,
		Status:            in.Status,
		BasePrice:         shared.Money{Amount: in.BasePrice, Currency: currency},
		MetalType:         in.MetalType,
		Karat:             in.Karat,
		GemstoneType:      in.GemstoneType,
		WeightGrams:       in.WeightGrams,
		PurityPercent:     in.PurityPercent,
		CertificateNumber: in.CertificateNumber,
		IsHandmade:        in.IsHandmade,
		IsFeatured:        in.IsFeatured,
		MetaTitle:         in.MetaTitle,
		MetaDescription:   in.MetaDescription,
	}
	if in.CompareAtPrice != nil {
		product.CompareAtPrice = &shared.Money{Amount: *in.CompareAtPrice, Currency: currency}
	}
	applyPublishTimestamp(product)

	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}
	return s.buildDetail(ctx, product)
}

func (s *Service) UpdateProduct(ctx context.Context, id shared.ID, in UpdateProductInput) (*domain.ProductDetail, error) {
	product, err := s.repo.FindProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.CategoryID != nil {
		if _, err := s.repo.FindCategoryByID(ctx, *in.CategoryID); err != nil {
			return nil, err
		}
		product.CategoryID = *in.CategoryID
	}
	if in.SKU != nil {
		sku := strings.TrimSpace(*in.SKU)
		if sku == "" {
			return nil, ErrSKURequired
		}
		if exists, err := s.repo.SKUExists(ctx, sku, &id); err != nil {
			return nil, err
		} else if exists {
			return nil, shared.ErrConflict
		}
		product.SKU = sku
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, ErrNameRequired
		}
		product.Name = name
	}
	if in.Slug != nil {
		slug := normalizeSlug(*in.Slug, product.Name)
		if exists, err := s.repo.SlugExists(ctx, slug, &id); err != nil {
			return nil, err
		} else if exists {
			return nil, shared.ErrConflict
		}
		product.Slug = slug
	}
	if in.Description != nil {
		product.Description = in.Description
	}
	if in.ShortDescription != nil {
		product.ShortDescription = in.ShortDescription
	}
	if in.JewelryType != nil {
		if !validJewelryType(*in.JewelryType) {
			return nil, ErrInvalidJewelry
		}
		product.JewelryType = *in.JewelryType
	}
	if in.Status != nil {
		if !validStatus(*in.Status) {
			return nil, ErrInvalidStatus
		}
		product.Status = *in.Status
		applyPublishTimestamp(product)
	}
	if in.BasePrice != nil {
		if *in.BasePrice < 0 {
			return nil, ErrNegativePrice
		}
		product.BasePrice.Amount = *in.BasePrice
	}
	if in.CompareAtPrice != nil {
		if *in.CompareAtPrice < 0 {
			return nil, ErrNegativePrice
		}
		product.CompareAtPrice = &shared.Money{Amount: *in.CompareAtPrice, Currency: product.BasePrice.Currency}
	}
	if in.MetalType != nil {
		if *in.MetalType != "" && !validMetalType(*in.MetalType) {
			return nil, ErrInvalidMetal
		}
		product.MetalType = in.MetalType
	}
	if in.Karat != nil {
		product.Karat = in.Karat
	}
	if in.GemstoneType != nil {
		if !validGemstoneType(*in.GemstoneType) {
			return nil, ErrInvalidGemstone
		}
		product.GemstoneType = *in.GemstoneType
	}
	if in.WeightGrams != nil {
		product.WeightGrams = in.WeightGrams
	}
	if in.PurityPercent != nil {
		product.PurityPercent = in.PurityPercent
	}
	if in.CertificateNumber != nil {
		product.CertificateNumber = in.CertificateNumber
	}
	if in.IsHandmade != nil {
		product.IsHandmade = *in.IsHandmade
	}
	if in.IsFeatured != nil {
		product.IsFeatured = *in.IsFeatured
	}
	if in.MetaTitle != nil {
		product.MetaTitle = in.MetaTitle
	}
	if in.MetaDescription != nil {
		product.MetaDescription = in.MetaDescription
	}

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return nil, err
	}
	return s.buildDetail(ctx, product)
}

func (s *Service) ArchiveProduct(ctx context.Context, id shared.ID) error {
	return s.repo.ArchiveProduct(ctx, id)
}

func (s *Service) CreateVariant(ctx context.Context, productID shared.ID, in CreateVariantInput) (*domain.VariantDetail, error) {
	if _, err := s.repo.FindProductByID(ctx, productID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrVariantRequired
	}
	sku := strings.TrimSpace(in.SKU)
	if sku == "" {
		return nil, ErrSKURequired
	}
	if exists, err := s.repo.VariantSKUExists(ctx, sku, nil); err != nil {
		return nil, err
	} else if exists {
		return nil, shared.ErrConflict
	}
	if in.InitialQuantity < 0 {
		return nil, ErrNegativeInventory
	}
	lowStock := in.LowStockThreshold
	if lowStock <= 0 {
		lowStock = 5
	}

	variant := &domain.ProductVariant{
		ID:              shared.ID(uuid.New()),
		ProductID:       productID,
		SKU:             sku,
		Name:            strings.TrimSpace(in.Name),
		SizeLabel:       in.SizeLabel,
		ColorLabel:      in.ColorLabel,
		PriceAdjustment: in.PriceAdjustment,
		WeightGrams:     in.WeightGrams,
		Barcode:         in.Barcode,
		IsDefault:       in.IsDefault,
		SortOrder:       in.SortOrder,
		IsActive:        true,
	}
	if err := s.repo.CreateVariant(ctx, variant); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertInventory(ctx, variant.ID, in.InitialQuantity, lowStock); err != nil {
		return nil, err
	}
	return s.repo.FindInventoryByVariantID(ctx, variant.ID)
}

func (s *Service) UpdateVariant(ctx context.Context, productID, variantID shared.ID, in UpdateVariantInput) (*domain.VariantDetail, error) {
	variant, err := s.repo.FindVariantByID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	if variant.ProductID != productID {
		return nil, shared.ErrNotFound
	}
	if in.SKU != nil {
		sku := strings.TrimSpace(*in.SKU)
		if sku == "" {
			return nil, ErrSKURequired
		}
		if exists, err := s.repo.VariantSKUExists(ctx, sku, &variantID); err != nil {
			return nil, err
		} else if exists {
			return nil, shared.ErrConflict
		}
		variant.SKU = sku
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, ErrVariantRequired
		}
		variant.Name = name
	}
	if in.SizeLabel != nil {
		variant.SizeLabel = in.SizeLabel
	}
	if in.ColorLabel != nil {
		variant.ColorLabel = in.ColorLabel
	}
	if in.PriceAdjustment != nil {
		variant.PriceAdjustment = *in.PriceAdjustment
	}
	if in.WeightGrams != nil {
		variant.WeightGrams = in.WeightGrams
	}
	if in.Barcode != nil {
		variant.Barcode = in.Barcode
	}
	if in.IsDefault != nil {
		variant.IsDefault = *in.IsDefault
	}
	if in.SortOrder != nil {
		variant.SortOrder = *in.SortOrder
	}
	if in.IsActive != nil {
		variant.IsActive = *in.IsActive
	}
	if err := s.repo.UpdateVariant(ctx, variant); err != nil {
		return nil, err
	}
	return s.repo.FindInventoryByVariantID(ctx, variantID)
}

func (s *Service) DeleteVariant(ctx context.Context, productID, variantID shared.ID) error {
	variant, err := s.repo.FindVariantByID(ctx, variantID)
	if err != nil {
		return err
	}
	if variant.ProductID != productID {
		return shared.ErrNotFound
	}
	return s.repo.DeactivateVariant(ctx, variantID)
}

func (s *Service) AddImage(ctx context.Context, productID shared.ID, in CreateImageInput) (*domain.ProductImage, error) {
	if _, err := s.repo.FindProductByID(ctx, productID); err != nil {
		return nil, err
	}
	url := strings.TrimSpace(in.URL)
	if url == "" {
		return nil, ErrImageURLRequired
	}
	if in.VariantID != nil {
		variant, err := s.repo.FindVariantByID(ctx, *in.VariantID)
		if err != nil {
			return nil, err
		}
		if variant.ProductID != productID {
			return nil, shared.ErrNotFound
		}
	}
	image := &domain.ProductImage{
		ID:        shared.ID(uuid.New()),
		ProductID: productID,
		VariantID: in.VariantID,
		URL:       url,
		AltText:   in.AltText,
		SortOrder: in.SortOrder,
		IsPrimary: in.IsPrimary,
	}
	if err := s.repo.CreateImage(ctx, image); err != nil {
		return nil, err
	}
	return image, nil
}

func (s *Service) DeleteImage(ctx context.Context, productID, imageID shared.ID) error {
	return s.repo.DeleteImage(ctx, productID, imageID)
}

func (s *Service) AdjustInventory(ctx context.Context, variantID shared.ID, in AdjustInventoryInput) (*domain.VariantDetail, error) {
	if in.Quantity < 0 {
		return nil, ErrNegativeInventory
	}
	detail, err := s.repo.FindInventoryByVariantID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	lowStock := detail.LowStockThreshold
	if in.LowStockThreshold != nil {
		lowStock = *in.LowStockThreshold
	}
	if err := s.repo.UpsertInventory(ctx, variantID, in.Quantity, lowStock); err != nil {
		return nil, err
	}
	return s.repo.FindInventoryByVariantID(ctx, variantID)
}

func (s *Service) buildDetail(ctx context.Context, product *domain.Product) (*domain.ProductDetail, error) {
	category, err := s.repo.FindCategoryByID(ctx, product.CategoryID)
	if err != nil {
		return nil, err
	}
	variants, err := s.repo.ListVariantDetailsByProductID(ctx, product.ID)
	if err != nil {
		return nil, err
	}
	images, err := s.repo.ListImagesByProductID(ctx, product.ID)
	if err != nil {
		return nil, err
	}
	return &domain.ProductDetail{
		Product:  *product,
		Category: *category,
		Variants: variants,
		Images:   images,
	}, nil
}

func validateCreateProduct(in CreateProductInput) error {
	if in.CategoryID == uuid.Nil {
		return ErrCategoryRequired
	}
	if strings.TrimSpace(in.SKU) == "" {
		return ErrSKURequired
	}
	if strings.TrimSpace(in.Name) == "" {
		return ErrNameRequired
	}
	if in.BasePrice < 0 {
		return ErrNegativePrice
	}
	if !validJewelryType(in.JewelryType) {
		return ErrInvalidJewelry
	}
	if !validStatus(in.Status) {
		return ErrInvalidStatus
	}
	if in.MetalType != nil && *in.MetalType != "" && !validMetalType(*in.MetalType) {
		return ErrInvalidMetal
	}
	if !validGemstoneType(in.GemstoneType) {
		return ErrInvalidGemstone
	}
	if in.CompareAtPrice != nil && *in.CompareAtPrice < 0 {
		return ErrNegativePrice
	}
	return nil
}

func applyPublishTimestamp(product *domain.Product) {
	if product.Status == domain.ProductStatusPublished && product.PublishedAt == nil {
		now := time.Now().UTC()
		product.PublishedAt = &now
	}
	if product.Status != domain.ProductStatusPublished {
		product.PublishedAt = nil
	}
}

func normalizeSlug(slug, name string) string {
	s := strings.TrimSpace(strings.ToLower(slug))
	if s == "" {
		s = strings.TrimSpace(strings.ToLower(name))
	}
	s = slugSanitizer.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func validStatus(s domain.ProductStatus) bool {
	switch s {
	case domain.ProductStatusDraft, domain.ProductStatusPublished, domain.ProductStatusArchived:
		return true
	default:
		return false
	}
}

func validJewelryType(t domain.JewelryType) bool {
	switch t {
	case domain.JewelryTypeRing, domain.JewelryTypeNecklace, domain.JewelryTypeBracelet,
		domain.JewelryTypeEarring, domain.JewelryTypePendant, domain.JewelryTypeAnklet,
		domain.JewelryTypeBrooch, domain.JewelryTypeSet, domain.JewelryTypeOther:
		return true
	default:
		return false
	}
}

func validMetalType(t domain.MetalType) bool {
	switch t {
	case domain.MetalTypeGoldYellow, domain.MetalTypeGoldWhite, domain.MetalTypeGoldRose,
		domain.MetalTypeSilver, domain.MetalTypePlatinum, domain.MetalTypeTitanium,
		domain.MetalTypeMixed, domain.MetalTypeOther:
		return true
	default:
		return false
	}
}

func validGemstoneType(t domain.GemstoneType) bool {
	switch t {
	case domain.GemstoneTypeNone, domain.GemstoneTypeDiamond, domain.GemstoneTypeRuby,
		domain.GemstoneTypeSapphire, domain.GemstoneTypeEmerald, domain.GemstoneTypePearl,
		domain.GemstoneTypeTurquoise, domain.GemstoneTypeAmethyst, domain.GemstoneTypeMixed,
		domain.GemstoneTypeOther:
		return true
	default:
		return false
	}
}
