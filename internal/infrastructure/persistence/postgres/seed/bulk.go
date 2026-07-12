package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type bulkCRMProfile struct {
	FirstName           string   `json:"firstName"`
	LastName            string   `json:"lastName"`
	Phone               string   `json:"phone"`
	Gender              string   `json:"gender,omitempty"`
	CustomerAgeRange    string   `json:"customerAgeRange"`
	CustomerType        string   `json:"customerType"`
	PurchasedCategories []string `json:"purchasedCategories"`
	FirstVisitDate      string   `json:"firstVisitDate"`
	Birthday            string   `json:"birthday"`
	MarriageDate        string   `json:"marriageDate,omitempty"`
}

func buildCRMProfileJSON(now time.Time, index int) (string, error) {
	first, last := pickName(index)
	p := bulkCRMProfile{
		FirstName:           first,
		LastName:            last,
		Phone:               bulkPhone(index),
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

func optionalBulkGender(gender string) any {
	if gender == "" {
		return nil
	}
	return gender
}

func parseBulkDate(value string) any {
	if value == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return t
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
	_ = roleCustomer
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

		customerRows := make([][]any, 0, end-start+1)

		for i := start; i <= end; i++ {
			registered := bulkRegisteredAt(now, i)
			profileJSON, err := buildCRMProfileJSON(now, i)
			if err != nil {
				return fmt.Errorf("bulk CRM profile %d: %w", i, err)
			}

			var p bulkCRMProfile
			if err := json.Unmarshal([]byte(profileJSON), &p); err != nil {
				return fmt.Errorf("bulk CRM profile %d: %w", i, err)
			}

			var email any
			if i%4 == 0 {
				email = bulkEmail(i)
			}

			customerRows = append(customerRows, []any{
				bulkCustomerID(i),
				p.FirstName,
				p.LastName,
				nil,
				p.Phone,
				email,
				nil,
				parseBulkDate(p.Birthday),
				parseBulkDate(p.MarriageDate),
				nil,
				parseBulkDate(p.FirstVisitDate),
				optionalBulkGender(p.Gender),
				p.CustomerType,
				p.CustomerAgeRange,
				p.PurchasedCategories,
				nil,
				nil,
				registered,
				registered,
			})
		}

		if err := copyRows(ctx, tx, "customers", []string{
			"id", "first_name", "last_name", "job", "phone", "email", "address",
			"birthday", "marriage_date", "important_date", "first_visit_date",
			"gender", "customer_type", "customer_age_range", "purchased_categories",
			"description", "signature_url", "created_at", "updated_at",
		}, customerRows); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("customers batch %d-%d: %w", start, end, err)
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

	roles, err := r.loadRoles(ctx)
	if err != nil {
		return err
	}

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
		var userStubRows [][]any

		for i := start; i <= end; i++ {
			if bulkHasWishlist(i) || bulkHasOrder(i) || bulkHasSecondOrder(i) {
				first, last := pickName(i)
				userStubRows = append(userStubRows, []any{
					bulkCustomerID(i), roles.customer, first, last, "active", now, now,
				})
			}

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

		if len(userStubRows) > 0 {
			if err := copyRows(ctx, tx, "users", []string{
				"id", "role_id", "first_name", "last_name", "status", "created_at", "updated_at",
			}, userStubRows); err != nil {
				tx.Rollback(ctx)
				return err
			}
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
	return nil
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
