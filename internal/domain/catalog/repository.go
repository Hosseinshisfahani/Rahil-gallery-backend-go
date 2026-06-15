package catalog

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Repository interface {
	// Categories
	FindCategoryByID(ctx context.Context, id shared.ID) (*Category, error)
	FindCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	ListActiveCategories(ctx context.Context) ([]Category, error)

	// Collections
	FindCollectionByID(ctx context.Context, id shared.ID) (*Collection, error)
	FindCollectionBySlug(ctx context.Context, slug string) (*Collection, error)
	ListActiveCollections(ctx context.Context) ([]Collection, error)

	// Products
	FindProductByID(ctx context.Context, id shared.ID) (*Product, error)
	FindProductBySlug(ctx context.Context, slug string) (*Product, error)
	FindPublishedProductBySlug(ctx context.Context, slug string) (*Product, error)
	ListAdmin(ctx context.Context, filter AdminListFilter, page, perPage int) (ListResult, error)
	ListPublished(ctx context.Context, filter PublicListFilter, page, perPage int) (ListResult, error)
	ListPublishedByCollection(ctx context.Context, collectionID shared.ID, page, perPage int) (ListResult, error)
	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, product *Product) error
	ArchiveProduct(ctx context.Context, id shared.ID) error
	SKUExists(ctx context.Context, sku string, excludeID *shared.ID) (bool, error)
	SlugExists(ctx context.Context, slug string, excludeID *shared.ID) (bool, error)

	// Variants
	FindVariantByID(ctx context.Context, id shared.ID) (*ProductVariant, error)
	FindVariantBySKU(ctx context.Context, sku string) (*ProductVariant, error)
	ListVariantsByProductID(ctx context.Context, productID shared.ID) ([]ProductVariant, error)
	CreateVariant(ctx context.Context, variant *ProductVariant) error
	UpdateVariant(ctx context.Context, variant *ProductVariant) error
	DeactivateVariant(ctx context.Context, id shared.ID) error
	VariantSKUExists(ctx context.Context, sku string, excludeID *shared.ID) (bool, error)

	// Images
	ListImagesByProductID(ctx context.Context, productID shared.ID) ([]ProductImage, error)
	CreateImage(ctx context.Context, image *ProductImage) error
	DeleteImage(ctx context.Context, productID, imageID shared.ID) error

	// Inventory
	FindInventoryByVariantID(ctx context.Context, variantID shared.ID) (*VariantDetail, error)
	ListVariantDetailsByProductID(ctx context.Context, productID shared.ID) ([]VariantDetail, error)
	UpsertInventory(ctx context.Context, variantID shared.ID, quantity, lowStockThreshold int) error
	AdjustInventory(ctx context.Context, variantID shared.ID, quantity int) error
}
