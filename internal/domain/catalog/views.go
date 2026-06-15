package catalog

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Availability string

const (
	AvailabilityInStock    Availability = "in_stock"
	AvailabilityOutOfStock Availability = "out_of_stock"
)

type AdminListFilter struct {
	Status       *ProductStatus
	CategoryID   *shared.ID
	JewelryType  *JewelryType
	MetalType    *MetalType
	GemstoneType *GemstoneType
	Featured     *bool
	Query        string
	PriceMin     *float64
	PriceMax     *float64
}

type PublicListFilter struct {
	CategorySlug   string
	CollectionSlug string
	JewelryType    *JewelryType
	MetalType      *MetalType
	GemstoneType   *GemstoneType
	Featured       bool
	Query          string
	PriceMin       *float64
	PriceMax       *float64
	Sort           string
}

type ListResult struct {
	Items   []ListItem
	Total   int
	Page    int
	PerPage int
}

type ListItem struct {
	Product
	CategoryName    string
	CategorySlug    string
	VariantCount    int
	PriceFrom       float64
	PriceTo         float64
	PrimaryImageURL *string
	Availability    Availability
}

type VariantDetail struct {
	ProductVariant
	Quantity          int
	ReservedQuantity  int
	LowStockThreshold int
}

type ProductDetail struct {
	Product
	Category Category
	Variants []VariantDetail
	Images   []ProductImage
}

type CollectionDetail struct {
	Collection
	ProductCount int
}

func (v VariantDetail) Available() int {
	return v.Quantity - v.ReservedQuantity
}

func (v VariantDetail) UnitPrice(base shared.Money) shared.Money {
	return shared.Money{
		Amount:   base.Amount + v.PriceAdjustment,
		Currency: base.Currency,
	}
}

func DeriveAvailability(variants []VariantDetail) Availability {
	for _, v := range variants {
		if v.IsActive && v.Available() > 0 {
			return AvailabilityInStock
		}
	}
	return AvailabilityOutOfStock
}

func PriceRange(base shared.Money, variants []VariantDetail) (from, to float64) {
	from = base.Amount
	to = base.Amount
	if len(variants) == 0 {
		return from, to
	}
	first := true
	for _, v := range variants {
		if !v.IsActive {
			continue
		}
		p := base.Amount + v.PriceAdjustment
		if first {
			from, to = p, p
			first = false
			continue
		}
		if p < from {
			from = p
		}
		if p > to {
			to = p
		}
	}
	return from, to
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
