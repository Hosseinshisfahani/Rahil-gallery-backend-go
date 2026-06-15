package seed

import (
	"context"
	"log"
)

func (r *Runner) resetBulkProducts(ctx context.Context) error {
	log.Printf("seed: clearing bulk products (SKU prefix %s)...", bulkProductSKUPrefix)

	stmts := []string{
		`DELETE FROM product_collections WHERE product_id IN (SELECT id FROM products WHERE sku LIKE $1)`,
		`DELETE FROM product_images WHERE product_id IN (SELECT id FROM products WHERE sku LIKE $1)`,
		`DELETE FROM inventory_items WHERE variant_id IN (SELECT id FROM product_variants WHERE sku LIKE $1)`,
		`DELETE FROM product_variants WHERE sku LIKE $1`,
		`DELETE FROM products WHERE sku LIKE $1`,
	}

	pattern := bulkProductSKUPrefix + "%"
	for _, q := range stmts {
		if _, err := r.pool.Exec(ctx, q, pattern); err != nil {
			return err
		}
	}
	return nil
}
