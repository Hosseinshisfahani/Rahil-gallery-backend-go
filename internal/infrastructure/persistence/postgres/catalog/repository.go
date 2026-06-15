package catalog

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindCategoryByID(ctx context.Context, id shared.ID) (*domain.Category, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, parent_id, name, slug, description, sort_order, is_active, created_at, updated_at
FROM categories WHERE id = $1`, id)
	c, err := scanCategory(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

func (r *Repository) FindCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, parent_id, name, slug, description, sort_order, is_active, created_at, updated_at
FROM categories WHERE slug = $1 AND is_active = TRUE`, slug)
	c, err := scanCategory(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

func (r *Repository) ListActiveCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, parent_id, name, slug, description, sort_order, is_active, created_at, updated_at
FROM categories WHERE is_active = TRUE ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) FindCollectionByID(ctx context.Context, id shared.ID) (*domain.Collection, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM collections WHERE id = $1`, id)
	c, err := scanCollection(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

func (r *Repository) FindCollectionBySlug(ctx context.Context, slug string) (*domain.Collection, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM collections WHERE slug = $1 AND is_active = TRUE`, slug)
	c, err := scanCollection(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

func (r *Repository) ListActiveCollections(ctx context.Context) ([]domain.Collection, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM collections WHERE is_active = TRUE ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Collection
	for rows.Next() {
		c, err := scanCollection(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) FindProductByID(ctx context.Context, id shared.ID) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx, `
SELECT `+productSelectCols+`
FROM products p WHERE p.id = $1 AND p.deleted_at IS NULL`, id)
	p, err := scanProduct(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

func (r *Repository) FindProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx, `
SELECT `+productSelectCols+`
FROM products p WHERE p.slug = $1 AND p.deleted_at IS NULL`, slug)
	p, err := scanProduct(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

func (r *Repository) FindPublishedProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx, `
SELECT `+productSelectCols+`
FROM products p
WHERE p.slug = $1 AND p.deleted_at IS NULL AND p.status = 'published'`, slug)
	p, err := scanProduct(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

func (r *Repository) ListAdmin(ctx context.Context, filter domain.AdminListFilter, page, perPage int) (domain.ListResult, error) {
	return r.listProducts(ctx, buildAdminListQuery(filter), page, perPage, false)
}

func (r *Repository) ListPublished(ctx context.Context, filter domain.PublicListFilter, page, perPage int) (domain.ListResult, error) {
	return r.listProducts(ctx, buildPublishedListQuery(filter), page, perPage, true)
}

func (r *Repository) ListPublishedByCollection(ctx context.Context, collectionID shared.ID, page, perPage int) (domain.ListResult, error) {
	return r.listProducts(ctx, buildCollectionListQuery(collectionID.String()), page, perPage, true)
}

func (r *Repository) listProducts(ctx context.Context, q listQuery, page, perPage int, public bool) (domain.ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 12
	}
	if perPage > 100 {
		perPage = 100
	}

	var total int
	if err := r.pool.QueryRow(ctx, q.countSQL(), q.args...).Scan(&total); err != nil {
		return domain.ListResult{}, err
	}

	totalPages := int(math.Max(1, math.Ceil(float64(total)/float64(perPage))))
	if page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * perPage

	listArgs := append(append([]any{}, q.args...), perPage, offset)
	rows, err := r.pool.Query(ctx, q.listSQL(perPage, offset), listArgs...)
	if err != nil {
		return domain.ListResult{}, err
	}
	defer rows.Close()

	var items []domain.ListItem
	for rows.Next() {
		item, err := scanListItem(rows)
		if err != nil {
			return domain.ListResult{}, err
		}
		if public {
			variants, err := r.ListVariantDetailsByProductID(ctx, item.ID)
			if err != nil {
				return domain.ListResult{}, err
			}
			item.Availability = domain.DeriveAvailability(variants)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, err
	}

	return domain.ListResult{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (r *Repository) CreateProduct(ctx context.Context, product *domain.Product) error {
	now := time.Now().UTC()
	product.CreatedAt = now
	product.UpdatedAt = now

	var compareAt *float64
	if product.CompareAtPrice != nil {
		compareAt = &product.CompareAtPrice.Amount
	}

	_, err := r.pool.Exec(ctx, `
INSERT INTO products (
	id, category_id, sku, name, slug, description, short_description,
	jewelry_type, status, base_price, compare_at_price, currency,
	metal_type, karat, gemstone_type, weight_grams, purity_percent,
	certificate_number, is_handmade, is_featured,
	meta_title, meta_description, published_at, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25
)`,
		product.ID, product.CategoryID, product.SKU, product.Name, product.Slug,
		product.Description, product.ShortDescription,
		product.JewelryType, product.Status,
		product.BasePrice.Amount, compareAt, product.BasePrice.Currency,
		product.MetalType, product.Karat, product.GemstoneType,
		product.WeightGrams, product.PurityPercent, product.CertificateNumber,
		product.IsHandmade, product.IsFeatured,
		product.MetaTitle, product.MetaDescription, product.PublishedAt,
		product.CreatedAt, product.UpdatedAt,
	)
	return mapInsertErr(err)
}

func (r *Repository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	product.UpdatedAt = time.Now().UTC()
	var compareAt *float64
	if product.CompareAtPrice != nil {
		compareAt = &product.CompareAtPrice.Amount
	}

	tag, err := r.pool.Exec(ctx, `
UPDATE products SET
	category_id = $2, sku = $3, name = $4, slug = $5,
	description = $6, short_description = $7,
	jewelry_type = $8, status = $9,
	base_price = $10, compare_at_price = $11, currency = $12,
	metal_type = $13, karat = $14, gemstone_type = $15,
	weight_grams = $16, purity_percent = $17, certificate_number = $18,
	is_handmade = $19, is_featured = $20,
	meta_title = $21, meta_description = $22, published_at = $23, updated_at = $24
WHERE id = $1 AND deleted_at IS NULL`,
		product.ID, product.CategoryID, product.SKU, product.Name, product.Slug,
		product.Description, product.ShortDescription,
		product.JewelryType, product.Status,
		product.BasePrice.Amount, compareAt, product.BasePrice.Currency,
		product.MetalType, product.Karat, product.GemstoneType,
		product.WeightGrams, product.PurityPercent, product.CertificateNumber,
		product.IsHandmade, product.IsFeatured,
		product.MetaTitle, product.MetaDescription, product.PublishedAt, product.UpdatedAt,
	)
	if err != nil {
		return mapInsertErr(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) ArchiveProduct(ctx context.Context, id shared.ID) error {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
UPDATE products SET status = 'archived', deleted_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL`, id, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) SKUExists(ctx context.Context, sku string, excludeID *shared.ID) (bool, error) {
	return r.exists(ctx, "SELECT 1 FROM products WHERE sku = $1 AND deleted_at IS NULL", sku, excludeID)
}

func (r *Repository) SlugExists(ctx context.Context, slug string, excludeID *shared.ID) (bool, error) {
	return r.exists(ctx, "SELECT 1 FROM products WHERE slug = $1 AND deleted_at IS NULL", slug, excludeID)
}

func (r *Repository) exists(ctx context.Context, baseSQL string, value string, excludeID *shared.ID) (bool, error) {
	sql := baseSQL
	args := []any{value}
	if excludeID != nil {
		sql += " AND id <> $2"
		args = append(args, *excludeID)
	}
	var one int
	err := r.pool.QueryRow(ctx, sql, args...).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) FindVariantByID(ctx context.Context, id shared.ID) (*domain.ProductVariant, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, product_id, sku, name, size_label, color_label, price_adjustment,
	weight_grams, barcode, is_default, sort_order, is_active, created_at, updated_at
FROM product_variants WHERE id = $1`, id)
	v, err := scanVariant(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &v, nil
}

func (r *Repository) FindVariantBySKU(ctx context.Context, sku string) (*domain.ProductVariant, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, product_id, sku, name, size_label, color_label, price_adjustment,
	weight_grams, barcode, is_default, sort_order, is_active, created_at, updated_at
FROM product_variants WHERE sku = $1`, sku)
	v, err := scanVariant(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &v, nil
}

func (r *Repository) ListVariantsByProductID(ctx context.Context, productID shared.ID) ([]domain.ProductVariant, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, product_id, sku, name, size_label, color_label, price_adjustment,
	weight_grams, barcode, is_default, sort_order, is_active, created_at, updated_at
FROM product_variants WHERE product_id = $1 ORDER BY sort_order ASC, created_at ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ProductVariant
	for rows.Next() {
		v, err := scanVariant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

func (r *Repository) CreateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	now := time.Now().UTC()
	variant.CreatedAt = now
	variant.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
INSERT INTO product_variants (
	id, product_id, sku, name, size_label, color_label, price_adjustment,
	weight_grams, barcode, is_default, sort_order, is_active, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		variant.ID, variant.ProductID, variant.SKU, variant.Name,
		variant.SizeLabel, variant.ColorLabel, variant.PriceAdjustment,
		variant.WeightGrams, variant.Barcode,
		variant.IsDefault, variant.SortOrder, variant.IsActive,
		variant.CreatedAt, variant.UpdatedAt,
	)
	return mapInsertErr(err)
}

func (r *Repository) UpdateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	variant.UpdatedAt = time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
UPDATE product_variants SET
	sku = $2, name = $3, size_label = $4, color_label = $5, price_adjustment = $6,
	weight_grams = $7, barcode = $8, is_default = $9, sort_order = $10, is_active = $11, updated_at = $12
WHERE id = $1`,
		variant.ID, variant.SKU, variant.Name,
		variant.SizeLabel, variant.ColorLabel, variant.PriceAdjustment,
		variant.WeightGrams, variant.Barcode,
		variant.IsDefault, variant.SortOrder, variant.IsActive, variant.UpdatedAt,
	)
	if err != nil {
		return mapInsertErr(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) DeactivateVariant(ctx context.Context, id shared.ID) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE product_variants SET is_active = FALSE, updated_at = $2 WHERE id = $1`, id, time.Now().UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) VariantSKUExists(ctx context.Context, sku string, excludeID *shared.ID) (bool, error) {
	return r.exists(ctx, "SELECT 1 FROM product_variants WHERE sku = $1", sku, excludeID)
}

func (r *Repository) ListImagesByProductID(ctx context.Context, productID shared.ID) ([]domain.ProductImage, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, product_id, variant_id, url, alt_text, sort_order, is_primary, created_at
FROM product_images WHERE product_id = $1
ORDER BY is_primary DESC, sort_order ASC, created_at ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ProductImage
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, img)
	}
	return items, rows.Err()
}

func (r *Repository) CreateImage(ctx context.Context, image *domain.ProductImage) error {
	image.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
INSERT INTO product_images (id, product_id, variant_id, url, alt_text, sort_order, is_primary, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		image.ID, image.ProductID, image.VariantID,
		image.URL, image.AltText, image.SortOrder, image.IsPrimary, image.CreatedAt,
	)
	return mapInsertErr(err)
}

func (r *Repository) DeleteImage(ctx context.Context, productID, imageID shared.ID) error {
	tag, err := r.pool.Exec(ctx, `
DELETE FROM product_images WHERE id = $1 AND product_id = $2`, imageID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) FindInventoryByVariantID(ctx context.Context, variantID shared.ID) (*domain.VariantDetail, error) {
	row := r.pool.QueryRow(ctx, `
SELECT pv.id, pv.product_id, pv.sku, pv.name, pv.size_label, pv.color_label, pv.price_adjustment,
	pv.weight_grams, pv.barcode, pv.is_default, pv.sort_order, pv.is_active, pv.created_at, pv.updated_at,
	COALESCE(ii.quantity, 0), COALESCE(ii.reserved_quantity, 0), COALESCE(ii.low_stock_threshold, 5)
FROM product_variants pv
LEFT JOIN inventory_items ii ON ii.variant_id = pv.id
WHERE pv.id = $1`, variantID)
	v, err := scanVariantDetail(row)
	if err != nil {
		return nil, mapErr(err)
	}
	return &v, nil
}

func (r *Repository) ListVariantDetailsByProductID(ctx context.Context, productID shared.ID) ([]domain.VariantDetail, error) {
	rows, err := r.pool.Query(ctx, `
SELECT pv.id, pv.product_id, pv.sku, pv.name, pv.size_label, pv.color_label, pv.price_adjustment,
	pv.weight_grams, pv.barcode, pv.is_default, pv.sort_order, pv.is_active, pv.created_at, pv.updated_at,
	COALESCE(ii.quantity, 0), COALESCE(ii.reserved_quantity, 0), COALESCE(ii.low_stock_threshold, 5)
FROM product_variants pv
LEFT JOIN inventory_items ii ON ii.variant_id = pv.id
WHERE pv.product_id = $1
ORDER BY pv.sort_order ASC, pv.created_at ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.VariantDetail
	for rows.Next() {
		v, err := scanVariantDetail(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertInventory(ctx context.Context, variantID shared.ID, quantity, lowStockThreshold int) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO inventory_items (variant_id, quantity, reserved_quantity, low_stock_threshold, updated_at)
VALUES ($1, $2, 0, $3, $4)
ON CONFLICT (variant_id) DO UPDATE SET
	quantity = EXCLUDED.quantity,
	low_stock_threshold = EXCLUDED.low_stock_threshold,
	updated_at = EXCLUDED.updated_at`,
		variantID, quantity, lowStockThreshold, time.Now().UTC(),
	)
	return err
}

func (r *Repository) AdjustInventory(ctx context.Context, variantID shared.ID, quantity int) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE inventory_items SET quantity = $2, updated_at = $3 WHERE variant_id = $1`,
		variantID, quantity, time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	return err
}

func mapInsertErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return shared.ErrConflict
	}
	if errors.As(err, &pgErr) && strings.Contains(pgErr.Message, "invalid input value for enum") {
		return fmt.Errorf("%w: %s", shared.ErrInvalidInput, pgErr.Message)
	}
	return err
}
