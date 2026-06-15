package catalog

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusPublished ProductStatus = "published"
	ProductStatusArchived  ProductStatus = "archived"
)

type JewelryType string

const (
	JewelryTypeRing      JewelryType = "ring"
	JewelryTypeNecklace  JewelryType = "necklace"
	JewelryTypeBracelet  JewelryType = "bracelet"
	JewelryTypeEarring   JewelryType = "earring"
	JewelryTypePendant   JewelryType = "pendant"
	JewelryTypeAnklet    JewelryType = "anklet"
	JewelryTypeBrooch    JewelryType = "brooch"
	JewelryTypeSet       JewelryType = "set"
	JewelryTypeOther     JewelryType = "other"
)

type MetalType string

const (
	MetalTypeGoldYellow MetalType = "gold_yellow"
	MetalTypeGoldWhite  MetalType = "gold_white"
	MetalTypeGoldRose   MetalType = "gold_rose"
	MetalTypeSilver     MetalType = "silver"
	MetalTypePlatinum   MetalType = "platinum"
	MetalTypeTitanium   MetalType = "titanium"
	MetalTypeMixed      MetalType = "mixed"
	MetalTypeOther      MetalType = "other"
)

type GemstoneType string

const (
	GemstoneTypeNone     GemstoneType = "none"
	GemstoneTypeDiamond  GemstoneType = "diamond"
	GemstoneTypeRuby     GemstoneType = "ruby"
	GemstoneTypeSapphire GemstoneType = "sapphire"
	GemstoneTypeEmerald  GemstoneType = "emerald"
	GemstoneTypePearl    GemstoneType = "pearl"
	GemstoneTypeTurquoise GemstoneType = "turquoise"
	GemstoneTypeAmethyst GemstoneType = "amethyst"
	GemstoneTypeMixed    GemstoneType = "mixed"
	GemstoneTypeOther    GemstoneType = "other"
)

type Category struct {
	ID          shared.ID
	ParentID    *shared.ID
	Name        string
	Slug        string
	Description *string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Collection struct {
	ID          shared.ID
	Name        string
	Slug        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Product struct {
	ID                shared.ID
	CategoryID        shared.ID
	SKU               string
	Name              string
	Slug              string
	Description       *string
	ShortDescription  *string
	JewelryType       JewelryType
	Status            ProductStatus
	BasePrice         shared.Money
	CompareAtPrice    *shared.Money
	MetalType         *MetalType
	Karat             *int16
	GemstoneType      GemstoneType
	WeightGrams       *float64
	PurityPercent     *float64
	CertificateNumber *string
	IsHandmade        bool
	IsFeatured        bool
	MetaTitle         *string
	MetaDescription   *string
	PublishedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type ProductVariant struct {
	ID              shared.ID
	ProductID       shared.ID
	SKU             string
	Name            string
	SizeLabel       *string
	ColorLabel      *string
	PriceAdjustment float64
	WeightGrams     *float64
	Barcode         *string
	IsDefault       bool
	SortOrder       int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProductImage struct {
	ID        shared.ID
	ProductID shared.ID
	VariantID *shared.ID
	URL       string
	AltText   *string
	SortOrder int
	IsPrimary bool
	CreatedAt time.Time
}
