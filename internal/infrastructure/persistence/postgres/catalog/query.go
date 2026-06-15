package catalog

import (
	"fmt"
	"strings"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
)

const productSelectCols = `
	p.id, p.category_id, p.sku, p.name, p.slug,
	p.description, p.short_description,
	p.jewelry_type, p.status,
	p.base_price, p.compare_at_price, p.currency,
	p.metal_type, p.karat, p.gemstone_type,
	p.weight_grams, p.purity_percent, p.certificate_number,
	p.is_handmade, p.is_featured,
	p.meta_title, p.meta_description,
	p.published_at, p.created_at, p.updated_at, p.deleted_at`

const listItemSelect = productSelectCols + `,
	c.name AS category_name,
	c.slug AS category_slug,
	COALESCE(vc.variant_count, 0) AS variant_count,
	COALESCE(vc.price_from, p.base_price) AS price_from,
	COALESCE(vc.price_to, p.base_price) AS price_to,
	pi.url AS primary_image`

const listFromJoins = `
FROM products p
JOIN categories c ON c.id = p.category_id
LEFT JOIN LATERAL (
	SELECT
		COUNT(*)::int AS variant_count,
		MIN(p.base_price + pv.price_adjustment) AS price_from,
		MAX(p.base_price + pv.price_adjustment) AS price_to
	FROM product_variants pv
	WHERE pv.product_id = p.id AND pv.is_active = TRUE
) vc ON TRUE
LEFT JOIN LATERAL (
	SELECT url
	FROM product_images
	WHERE product_id = p.id
	ORDER BY is_primary DESC, sort_order ASC, created_at ASC
	LIMIT 1
) pi ON TRUE`

type listQuery struct {
	where  string
	args   []any
	order  string
	joins  string
}

func buildAdminListQuery(filter domain.AdminListFilter) listQuery {
	q := listQuery{
		where: "p.deleted_at IS NULL",
		order: "p.updated_at DESC",
	}
	if filter.Status != nil {
		q.args = append(q.args, string(*filter.Status))
		q.where += fmt.Sprintf(" AND p.status = $%d", len(q.args))
	}
	if filter.CategoryID != nil {
		q.args = append(q.args, *filter.CategoryID)
		q.where += fmt.Sprintf(" AND p.category_id = $%d", len(q.args))
	}
	if filter.JewelryType != nil {
		q.args = append(q.args, string(*filter.JewelryType))
		q.where += fmt.Sprintf(" AND p.jewelry_type = $%d", len(q.args))
	}
	if filter.MetalType != nil {
		q.args = append(q.args, string(*filter.MetalType))
		q.where += fmt.Sprintf(" AND p.metal_type = $%d", len(q.args))
	}
	if filter.GemstoneType != nil {
		q.args = append(q.args, string(*filter.GemstoneType))
		q.where += fmt.Sprintf(" AND p.gemstone_type = $%d", len(q.args))
	}
	if filter.Featured != nil {
		q.args = append(q.args, *filter.Featured)
		q.where += fmt.Sprintf(" AND p.is_featured = $%d", len(q.args))
	}
	if filter.PriceMin != nil {
		q.args = append(q.args, *filter.PriceMin)
		q.where += fmt.Sprintf(" AND COALESCE(vc.price_from, p.base_price) >= $%d", len(q.args))
	}
	if filter.PriceMax != nil {
		q.args = append(q.args, *filter.PriceMax)
		q.where += fmt.Sprintf(" AND COALESCE(vc.price_to, p.base_price) <= $%d", len(q.args))
	}
	if text := strings.TrimSpace(filter.Query); text != "" {
		q.args = append(q.args, "%"+text+"%")
		n := len(q.args)
		q.where += fmt.Sprintf(
			" AND (p.name ILIKE $%d OR p.sku ILIKE $%d OR p.slug ILIKE $%d)",
			n, n, n,
		)
	}
	return q
}

func buildPublishedListQuery(filter domain.PublicListFilter) listQuery {
	q := listQuery{
		where: "p.deleted_at IS NULL AND p.status = 'published'",
		order: publishedSort(filter.Sort),
	}
	if slug := strings.TrimSpace(filter.CategorySlug); slug != "" {
		q.args = append(q.args, slug)
		q.where += fmt.Sprintf(" AND c.slug = $%d", len(q.args))
	}
	if filter.JewelryType != nil {
		q.args = append(q.args, string(*filter.JewelryType))
		q.where += fmt.Sprintf(" AND p.jewelry_type = $%d", len(q.args))
	}
	if filter.MetalType != nil {
		q.args = append(q.args, string(*filter.MetalType))
		q.where += fmt.Sprintf(" AND p.metal_type = $%d", len(q.args))
	}
	if filter.GemstoneType != nil {
		q.args = append(q.args, string(*filter.GemstoneType))
		q.where += fmt.Sprintf(" AND p.gemstone_type = $%d", len(q.args))
	}
	if filter.Featured {
		q.where += " AND p.is_featured = TRUE"
	}
	if filter.PriceMin != nil {
		q.args = append(q.args, *filter.PriceMin)
		q.where += fmt.Sprintf(" AND COALESCE(vc.price_from, p.base_price) >= $%d", len(q.args))
	}
	if filter.PriceMax != nil {
		q.args = append(q.args, *filter.PriceMax)
		q.where += fmt.Sprintf(" AND COALESCE(vc.price_to, p.base_price) <= $%d", len(q.args))
	}
	if text := strings.TrimSpace(filter.Query); text != "" {
		q.args = append(q.args, "%"+text+"%")
		n := len(q.args)
		q.where += fmt.Sprintf(
			" AND (p.name ILIKE $%d OR p.short_description ILIKE $%d)",
			n, n,
		)
	}
	if slug := strings.TrimSpace(filter.CollectionSlug); slug != "" {
		q.joins = `
JOIN product_collections pc ON pc.product_id = p.id
JOIN collections col ON col.id = pc.collection_id AND col.is_active = TRUE`
		q.args = append(q.args, slug)
		q.where += fmt.Sprintf(" AND col.slug = $%d", len(q.args))
	}
	return q
}

func buildCollectionListQuery(collectionID string) listQuery {
	q := listQuery{
		where: "p.deleted_at IS NULL AND p.status = 'published'",
		joins: `
JOIN product_collections pc ON pc.product_id = p.id`,
		order: "p.is_featured DESC, p.published_at DESC NULLS LAST, p.created_at DESC",
	}
	q.args = append(q.args, collectionID)
	q.where += fmt.Sprintf(" AND pc.collection_id = $%d", len(q.args))
	return q
}

func publishedSort(sort string) string {
	switch strings.TrimSpace(sort) {
	case "price_asc":
		return "COALESCE(vc.price_from, p.base_price) ASC"
	case "price_desc":
		return "COALESCE(vc.price_to, p.base_price) DESC"
	case "featured":
		return "p.is_featured DESC, p.published_at DESC NULLS LAST"
	default:
		return "p.published_at DESC NULLS LAST, p.created_at DESC"
	}
}

func (q listQuery) countSQL() string {
	return fmt.Sprintf("SELECT COUNT(*) %s %s WHERE %s", listFromJoins, q.joins, q.where)
}

func (q listQuery) listSQL(limit, offset int) string {
	args := append([]any{}, q.args...)
	args = append(args, limit, offset)
	limitIdx := len(q.args) + 1
	offsetIdx := len(q.args) + 2
	return fmt.Sprintf(
		"SELECT %s %s %s WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d",
		listItemSelect, listFromJoins, q.joins, q.where, q.order, limitIdx, offsetIdx,
	)
}
