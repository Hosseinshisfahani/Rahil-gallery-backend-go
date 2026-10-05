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

func copyRows(ctx context.Context, tx pgx.Tx, table string, columns []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := tx.CopyFrom(ctx, pgx.Identifier{table}, columns, pgx.CopyFromRows(rows))
	return err
}
