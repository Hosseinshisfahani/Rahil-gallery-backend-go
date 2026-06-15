package seed

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (r *Runner) seedBulkProducts(ctx context.Context) error {
	count := r.opts.bulkProductCount()
	if count == 0 {
		return nil
	}

	log.Printf("seed: inserting %d bulk products (batches of %d)...", count, batchSize)
	now := time.Now().UTC()

	for start := 1; start <= count; start += batchSize {
		end := start + batchSize - 1
		if end > count {
			end = count
		}

		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return err
		}

		productRows := make([][]any, 0, end-start+1)
		variantRows := make([][]any, 0, (end-start+1)*2)
		inventoryRows := make([][]any, 0, (end-start+1)*2)
		imageRows := make([][]any, 0, end-start+1)
		collectionRows := make([][]any, 0, end-start+1)

		for i := start; i <= end; i++ {
			name, slug := bulkProductName(i)
			basePrice := bulkBasePrice(i)
			status := bulkProductStatus(i)
			jt := bulkJewelryType(i)
			metal := bulkMetalType(i)
			gemstone := bulkGemstoneType(i)
			karat := bulkKarat(i)
			weight := bulkWeightGrams(i)
			compareAt := bulkCompareAtPrice(i, basePrice)
			publishedAt := bulkPublishedAt(now, i)

			productRows = append(productRows, []any{
				bulkProductID(i), bulkCategoryID(i), bulkProductSKU(i), name, slug,
				fmt.Sprintf("%s — a %s piece from the Rahil Gallery seed catalog.", name, jt),
				fmt.Sprintf("Seed catalog %s", jt),
				jt, status, basePrice, compareAt, "IRR",
				metal, karat, gemstone, weight, nil, nil,
				bulkIsHandmade(i), bulkIsFeatured(i),
				nil, nil, publishedAt, now, now,
			})

			for v := 0; v < bulkVariantCount(i); v++ {
				variantName, sizeLabel := bulkVariantLabel(jt, v)
				variantRows = append(variantRows, []any{
					bulkVariantID(i, v), bulkProductID(i), bulkVariantSKU(i, v), variantName,
					sizeLabel, nil, bulkVariantPriceAdjustment(v), nil, nil,
					v == 0, v, true, now, now,
				})
				inventoryRows = append(inventoryRows, []any{
					bulkVariantID(i, v), bulkInventoryQuantity(i, v), 0,
					bulkLowStockThreshold(i), now,
				})
			}

			imageRows = append(imageRows, []any{
				bulkProductImageID(i), bulkProductID(i), nil,
				bulkImageURL(i), name, 0, true, now,
			})

			if colID := bulkCollectionForProduct(i); colID != nil {
				collectionRows = append(collectionRows, []any{bulkProductID(i), *colID})
			}
		}

		if err := copyRows(ctx, tx, "products", []string{
			"id", "category_id", "sku", "name", "slug", "description", "short_description",
			"jewelry_type", "status", "base_price", "compare_at_price", "currency",
			"metal_type", "karat", "gemstone_type", "weight_grams", "purity_percent", "certificate_number",
			"is_handmade", "is_featured", "meta_title", "meta_description", "published_at", "created_at", "updated_at",
		}, productRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("products batch %d-%d: %w", start, end, err)
		}

		if err := copyRows(ctx, tx, "product_variants", []string{
			"id", "product_id", "sku", "name", "size_label", "color_label", "price_adjustment",
			"weight_grams", "barcode", "is_default", "sort_order", "is_active", "created_at", "updated_at",
		}, variantRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("variants batch %d-%d: %w", start, end, err)
		}

		if err := copyRows(ctx, tx, "inventory_items", []string{
			"variant_id", "quantity", "reserved_quantity", "low_stock_threshold", "updated_at",
		}, inventoryRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("inventory batch %d-%d: %w", start, end, err)
		}

		if err := copyRows(ctx, tx, "product_images", []string{
			"id", "product_id", "variant_id", "url", "alt_text", "sort_order", "is_primary", "created_at",
		}, imageRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("images batch %d-%d: %w", start, end, err)
		}

		if len(collectionRows) > 0 {
			if err := copyRows(ctx, tx, "product_collections", []string{
				"product_id", "collection_id",
			}, collectionRows); err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("product_collections batch %d-%d: %w", start, end, err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		if end == count || end%1000 == 0 {
			log.Printf("seed:   products %d / %d", end, count)
		}
	}

	return nil
}
