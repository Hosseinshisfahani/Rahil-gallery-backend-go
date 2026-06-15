package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type bulkCRMProfile struct {
	Gender              string   `json:"gender,omitempty"`
	CustomerAgeRange    string   `json:"customerAgeRange"`
	CustomerType        string   `json:"customerType"`
	PurchasedCategories []string `json:"purchasedCategories"`
	FirstVisitDate      string   `json:"firstVisitDate"`
	Birthday            string   `json:"birthday"`
	MarriageDate        string   `json:"marriageDate,omitempty"`
}

func buildCRMProfileJSON(now time.Time, index int) (string, error) {
	p := bulkCRMProfile{
		Gender:              bulkCRMGender(index),
		CustomerAgeRange:    bulkCRMAgeRange(index),
		CustomerType:        bulkCRMCustomerType(index),
		PurchasedCategories: bulkCRMCategories(index),
		FirstVisitDate:      bulkCRMFirstVisit(now, index),
		Birthday:            bulkCRMBirthday(index),
		MarriageDate:        bulkCRMMarriageDate(index),
	}
	b, err := json.Marshal(p)
	return string(b), err
}

func (r *Runner) refreshCommerceStats(ctx context.Context) error {
	log.Println("seed: refreshing customer commerce stats...")
	_, err := r.pool.Exec(ctx, `SELECT refresh_all_customer_commerce_stats()`)
	return err
}

var variantCatalog = []struct {
	id, productName, variantName, sku string
	price                             float64
}{
	{IDVariantSolitaire14.String(), "Solitaire Diamond Ring", "Size 14", "RG-SOL-001-14", 185_000_000},
	{IDVariantSolitaire16.String(), "Solitaire Diamond Ring", "Size 16", "RG-SOL-001-16", 190_000_000},
	{IDVariantEternity12.String(), "Eternity Band", "Size 12", "RG-ETR-001-12", 92_000_000},
	{IDVariantPearlStd.String(), "Pearl Cocktail Ring", "Standard", "RG-PRL-001-STD", 48_000_000},
}

func (r *Runner) seedBulkCustomers(ctx context.Context, roleCustomer string) error {
	count := r.opts.bulkCustomerCount()
	if count == 0 {
		return nil
	}

	log.Printf("seed: inserting %d bulk customers (batches of %d)...", count, batchSize)
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

		userRows := make([][]any, 0, end-start+1)
		profileRows := make([][]any, 0, end-start+1)

		for i := start; i <= end; i++ {
			first, last := pickName(i)
			registered := bulkRegisteredAt(now, i)
			lastActivity := bulkLastActivity(now, i)
			isVIP := bulkIsVIP(i)
			tags := formatPGTextArray(bulkTags(i))

			var email any
			if i%4 == 0 {
				email = bulkEmail(i)
			}

			var vipSource any
			if isVIP {
				vipSource = "manual"
			}

			userRows = append(userRows, []any{
				bulkCustomerID(i), roleCustomer, email, bulkPhone(i), nil,
				first, last, bulkUserStatus(i), registered, registered,
			})

			var importMode, importProfile any
			if bulkHasCRMProfile(i) {
				profileJSON, err := buildCRMProfileJSON(now, i)
				if err == nil {
					importMode = "history_included"
					importProfile = profileJSON
				}
			}

			profileRows = append(profileRows, []any{
				bulkCustomerID(i), bulkLocale(i), nil, isVIP, vipSource,
				importMode, importProfile, tags, nil, nil, lastActivity, registered, registered,
			})
		}

		if err := copyRows(ctx, tx, "users", []string{
			"id", "role_id", "email", "phone", "password_hash",
			"first_name", "last_name", "status", "created_at", "updated_at",
		}, userRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("users batch %d-%d: %w", start, end, err)
		}

		if err := copyRows(ctx, tx, "customer_profiles", []string{
			"user_id", "locale", "default_ring_size", "is_vip", "vip_source",
			"import_mode", "import_profile", "tags", "block_reason", "block_note",
			"last_activity_at", "created_at", "updated_at",
		}, profileRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("profiles batch %d-%d: %w", start, end, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		if end == count || end%10_000 == 0 {
			log.Printf("seed:   customers %d / %d", end, count)
		}
	}

	return nil
}

func (r *Runner) seedBulkCommerce(ctx context.Context) error {
	count := r.opts.bulkCustomerCount()
	if count == 0 {
		return nil
	}

	orderCount := 0
	for i := 1; i <= count; i++ {
		if bulkHasOrder(i) {
			orderCount++
		}
		if bulkHasSecondOrder(i) {
			orderCount++
		}
	}

	log.Printf("seed: inserting ~%d bulk orders and wishlist items...", orderCount)

	if _, err := r.pool.Exec(ctx, `ALTER TABLE orders DISABLE TRIGGER trg_orders_sync_commerce_stats`); err != nil {
		return fmt.Errorf("disable commerce trigger: %w", err)
	}
	defer func() {
		_, _ = r.pool.Exec(context.Background(), `ALTER TABLE orders ENABLE TRIGGER trg_orders_sync_commerce_stats`)
	}()

	now := time.Now().UTC()
	orderSeq := 0

	for start := 1; start <= count; start += batchSize {
		end := start + batchSize - 1
		if end > count {
			end = count
		}

		var orderRows [][]any
		var itemRows [][]any
		var wishlistRows [][]any

		for i := start; i <= end; i++ {
			if bulkHasWishlist(i) {
				v := variantCatalog[bulkVariantIndex(i)]
				wishlistRows = append(wishlistRows, []any{
					bulkCustomerID(i), v.id, bulkLastActivity(now, i),
				})
			}

			if bulkHasOrder(i) {
				orderSeq++
				orderRows, itemRows = appendBulkOrder(orderRows, itemRows, i, orderSeq, now, false)
			}
			if bulkHasSecondOrder(i) {
				orderSeq++
				orderRows, itemRows = appendBulkOrder(orderRows, itemRows, i, orderSeq, now, true)
			}
		}

		if len(orderRows) == 0 && len(wishlistRows) == 0 {
			continue
		}

		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return err
		}

		if len(orderRows) > 0 {
			if err := copyRows(ctx, tx, "orders", []string{
				"id", "order_number", "user_id", "status", "subtotal",
				"discount_amount", "shipping_amount", "tax_amount", "total_amount",
				"currency", "placed_at", "created_at", "updated_at",
			}, orderRows); err != nil {
				tx.Rollback(ctx)
				return err
			}
			if err := copyRows(ctx, tx, "order_items", []string{
				"id", "order_id", "variant_id", "product_name", "variant_name", "sku",
				"jewelry_type", "metal_type", "karat", "quantity", "unit_price", "line_total",
			}, itemRows); err != nil {
				tx.Rollback(ctx)
				return err
			}
		}

		if len(wishlistRows) > 0 {
			if err := copyRows(ctx, tx, "wishlist_items", []string{
				"user_id", "variant_id", "created_at",
			}, wishlistRows); err != nil {
				tx.Rollback(ctx)
				return err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	log.Printf("seed:   orders inserted: %d", orderSeq)
	return r.refreshCommerceStats(ctx)
}

func appendBulkOrder(orders, items [][]any, customerIndex, orderSeq int, now time.Time, second bool) ([][]any, [][]any) {
	vi := bulkVariantIndex(customerIndex)
	if second {
		vi = (vi + 1) % len(variantCatalog)
	}
	v := variantCatalog[vi]
	total := bulkOrderTotal(customerIndex + orderSeq)
	placed := now.AddDate(0, 0, -(customerIndex%365))
	orderID := bulkOrderID(orderSeq)
	orderNum := fmt.Sprintf("RG-BULK-%08d", orderSeq)

	orders = append(orders, []any{
		orderID, orderNum, bulkCustomerID(customerIndex), bulkOrderStatus(customerIndex),
		total, 0, 0, 0, total, "IRR", placed, placed, placed,
	})

	items = append(items, []any{
		bulkOrderItemID(orderSeq), orderID, v.id, v.productName, v.variantName, v.sku,
		"ring", "gold_white", 18, 1, total, total,
	})

	return orders, items
}

func copyRows(ctx context.Context, tx pgx.Tx, table string, columns []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := tx.CopyFrom(ctx, pgx.Identifier{table}, columns, pgx.CopyFromRows(rows))
	return err
}

func formatPGTextArray(values []string) string {
	if len(values) == 0 {
		return "{}"
	}
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(quoted, ",") + "}"
}
