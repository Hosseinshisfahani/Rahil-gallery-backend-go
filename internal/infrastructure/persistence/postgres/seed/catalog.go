package seed

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Runner) seedCatalog(ctx context.Context, tx pgx.Tx) error {
	now := time.Now().UTC()

	for _, cat := range fixtureCategories {
		if err := exec(ctx, tx, `
INSERT INTO categories (id, name, slug, description, sort_order, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, TRUE, $6, $6)
ON CONFLICT (id) DO NOTHING`,
			cat.id, cat.name, cat.slug, cat.description, cat.sort, now,
		); err != nil {
			return err
		}
	}

	for _, col := range fixtureCollections {
		if err := exec(ctx, tx, `
INSERT INTO collections (id, name, slug, description, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, TRUE, $5, $5)
ON CONFLICT (id) DO NOTHING`,
			col.id, col.name, col.slug, col.description, now,
		); err != nil {
			return err
		}
	}

	products := []struct {
		id, sku, name, slug, categoryID string
		jewelryType                       string
		basePrice                         float64
		metal                             string
		karat                             int
		gemstone                          string
		collectionID                      *string
	}{
		{IDProductSolitaire.String(), "RG-SOL-001", "Solitaire Diamond Ring", "solitaire-diamond-ring", IDCategoryRings.String(), "ring", 185_000_000, "gold_white", 18, "diamond", strPtr(IDCollectionBridal.String())},
		{IDProductEternity.String(), "RG-ETR-001", "Eternity Band", "eternity-band", IDCategoryRings.String(), "ring", 92_000_000, "gold_yellow", 18, "diamond", strPtr(IDCollectionBridal.String())},
		{IDProductPearl.String(), "RG-PRL-001", "Pearl Cocktail Ring", "pearl-cocktail-ring", IDCategoryRings.String(), "ring", 48_000_000, "gold_rose", 18, "pearl", strPtr(IDCollectionEveryday.String())},
	}

	for _, p := range products {
		if err := exec(ctx, tx, `
INSERT INTO products (
	id, category_id, sku, name, slug, short_description, jewelry_type, status, base_price, currency,
	metal_type, karat, gemstone_type, is_featured, is_handmade, published_at, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,'published',$8,'IRR',$9,$10,$11,TRUE,FALSE,$12,$12,$12)
ON CONFLICT (id) DO NOTHING`,
			p.id, p.categoryID, p.sku, p.name, p.slug,
			"Handcrafted piece from the Rahil Gallery atelier.",
			p.jewelryType, p.basePrice, p.metal, p.karat, p.gemstone, now,
		); err != nil {
			return err
		}
		if p.collectionID != nil {
			if err := exec(ctx, tx, `
INSERT INTO product_collections (product_id, collection_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING`,
				p.id, *p.collectionID,
			); err != nil {
				return err
			}
		}
	}

	variants := []struct {
		id, productID, sku, name, size string
		adjustment                     float64
	}{
		{IDVariantSolitaire14.String(), IDProductSolitaire.String(), "RG-SOL-001-14", "Size 14", "14", 0},
		{IDVariantSolitaire16.String(), IDProductSolitaire.String(), "RG-SOL-001-16", "Size 16", "16", 5_000_000},
		{IDVariantEternity12.String(), IDProductEternity.String(), "RG-ETR-001-12", "Size 12", "12", 0},
		{IDVariantPearlStd.String(), IDProductPearl.String(), "RG-PRL-001-STD", "Standard", "", 0},
	}

	for _, v := range variants {
		if err := exec(ctx, tx, `
INSERT INTO product_variants (
	id, product_id, sku, name, size_label, price_adjustment, is_default, sort_order, is_active, created_at, updated_at
) VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,TRUE,0,TRUE,$7,$7)
ON CONFLICT (id) DO NOTHING`,
			v.id, v.productID, v.sku, v.name, v.size, v.adjustment, now,
		); err != nil {
			return err
		}

		if err := exec(ctx, tx, `
INSERT INTO inventory_items (variant_id, quantity, reserved_quantity, low_stock_threshold, updated_at)
VALUES ($1, 25, 0, 3, $2)
ON CONFLICT (variant_id) DO NOTHING`,
			v.id, now,
		); err != nil {
			return err
		}
	}

	images := []struct {
		id, productID, alt string
		jewelryType       string
		slot              int
	}{
		{IDImageSolitaire.String(), IDProductSolitaire.String(), "Solitaire diamond ring", "ring", 0},
		{IDImageEternity.String(), IDProductEternity.String(), "Eternity band", "ring", 1},
		{IDImagePearl.String(), IDProductPearl.String(), "Pearl cocktail ring", "ring", 2},
	}
	for _, img := range images {
		url := fixtureProductImageURL(img.jewelryType, img.slot)
		if err := exec(ctx, tx, `
INSERT INTO product_images (id, product_id, url, alt_text, sort_order, is_primary, created_at)
VALUES ($1, $2, $3, $4, 0, TRUE, $5)
ON CONFLICT (id) DO UPDATE SET url = EXCLUDED.url, alt_text = EXCLUDED.alt_text`,
			img.id, img.productID, url, img.alt, now,
		); err != nil {
			return err
		}
	}

	return nil
}
