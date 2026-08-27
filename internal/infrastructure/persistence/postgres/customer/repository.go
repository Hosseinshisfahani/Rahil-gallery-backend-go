package customer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter, page, perPage int) (domain.ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	q := buildListQuery(filter, nil)

	if filter.SkipCount {
		return r.listWithoutCount(ctx, q, page, perPage)
	}

	total, err := r.countCustomers(ctx, q)
	if err != nil {
		return domain.ListResult{}, err
	}

	totalPages := int(math.Max(1, math.Ceil(float64(total)/float64(perPage))))
	if page > totalPages && total > 0 {
		page = totalPages
	}
	offset := (page - 1) * perPage

	items, err := r.fetchCustomerPage(ctx, q, perPage, offset)
	if err != nil {
		return domain.ListResult{}, err
	}

	return domain.ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		TotalExact: true,
	}, nil
}

func (r *Repository) listWithoutCount(ctx context.Context, q listQuery, page, perPage int) (domain.ListResult, error) {
	offset := (page - 1) * perPage
	items, err := r.fetchCustomerPage(ctx, q, perPage+1, offset)
	if err != nil {
		return domain.ListResult{}, err
	}

	hasMore := len(items) > perPage
	if hasMore {
		items = items[:perPage]
	}

	return domain.ListResult{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		HasMore:    hasMore,
		TotalExact: false,
	}, nil
}

func (r *Repository) countCustomers(ctx context.Context, q listQuery) (int, error) {
	countQ := fmt.Sprintf(`SELECT COUNT(*) %s %s`, listBaseFrom, q.where)
	var total int
	if err := r.pool.QueryRow(ctx, countQ, q.args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) fetchCustomerPage(ctx context.Context, q listQuery, limit, offset int) ([]domain.ListRow, error) {
	listQ := fmt.Sprintf(`%s %s %s ORDER BY c.created_at DESC LIMIT %d OFFSET %d`,
		listSelectColumns, listBaseFrom, q.where, limit, offset)

	rows, err := r.pool.Query(ctx, listQ, q.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ListRow
	for rows.Next() {
		var row domain.ListRow
		if err := rows.Scan(
			&row.ID,
			&row.FullName,
			&row.Phone,
			&row.Email,
			&row.CustomerType,
			&row.CustomerAgeRange,
			&row.Gender,
			&row.PurchasedCategories,
			&row.CreatedAt,
		); err != nil {
			return nil, err
		}
		domain.NormalizeListRow(&row)
		row.PurchasedCategories = domain.StringSliceOrEmpty(row.PurchasedCategories)
		items = append(items, row)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id shared.ID) (*domain.Customer, error) {
	const q = `
SELECT
  id, first_name, last_name, job, phone, email, address,
  melli_code, postal_code,
  birthday, marriage_date, important_date, first_visit_date,
  gender, customer_type, customer_age_range, purchased_categories,
  description, marketer_note, signature_url, created_at, updated_at
FROM customers
WHERE id = $1 AND deleted_at IS NULL`

	row := r.pool.QueryRow(ctx, q, id)
	customer, err := scanCustomer(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return customer, nil
}

func (r *Repository) PhoneExists(ctx context.Context, phone string, excludeID *shared.ID) (bool, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return false, nil
	}

	var exists bool
	if excludeID != nil {
		err := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM customers WHERE phone = $1 AND deleted_at IS NULL AND id <> $2)`,
			phone, *excludeID,
		).Scan(&exists)
		return exists, err
	}

	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM customers WHERE phone = $1 AND deleted_at IS NULL)`,
		phone,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) EmailExists(ctx context.Context, email string, excludeID *shared.ID) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, nil
	}

	var exists bool
	if excludeID != nil {
		err := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM customers WHERE lower(email) = $1 AND deleted_at IS NULL AND id <> $2)`,
			email, *excludeID,
		).Scan(&exists)
		return exists, err
	}

	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM customers WHERE lower(email) = $1 AND deleted_at IS NULL)`,
		email,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) Create(ctx context.Context, customer *domain.Customer) error {
	const q = `
INSERT INTO customers (
  id, first_name, last_name, job, phone, email, address,
  melli_code, postal_code,
  birthday, marriage_date, important_date, first_visit_date,
  gender, customer_type, customer_age_range, purchased_categories,
  description, marketer_note, signature_url, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9,
  $10, $11, $12, $13,
  $14, $15, $16, $17,
  $18, $19, $20, $21, $22
)`

	_, err := r.pool.Exec(ctx, q,
		customer.ID,
		customer.FirstName,
		customer.LastName,
		customer.Job,
		customer.Phone,
		customer.Email,
		customer.Address,
		customer.MelliCode,
		customer.PostalCode,
		customer.Birthday,
		customer.MarriageDate,
		customer.ImportantDate,
		customer.FirstVisitDate,
		customer.Gender,
		customer.CustomerType,
		customer.CustomerAgeRange,
		customer.PurchasedCategories,
		customer.Description,
		customer.MarketerNote,
		customer.SignatureURL,
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	return err
}

func (r *Repository) Update(ctx context.Context, customer *domain.Customer) error {
	const q = `
UPDATE customers SET
  first_name = $2,
  last_name = $3,
  job = $4,
  phone = $5,
  email = $6,
  address = $7,
  melli_code = $8,
  postal_code = $9,
  birthday = $10,
  marriage_date = $11,
  important_date = $12,
  first_visit_date = $13,
  gender = $14,
  customer_type = $15,
  customer_age_range = $16,
  purchased_categories = $17,
  description = $18,
  marketer_note = $19,
  signature_url = $20,
  updated_at = $21
WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, q,
		customer.ID,
		customer.FirstName,
		customer.LastName,
		customer.Job,
		customer.Phone,
		customer.Email,
		customer.Address,
		customer.MelliCode,
		customer.PostalCode,
		customer.Birthday,
		customer.MarriageDate,
		customer.ImportantDate,
		customer.FirstVisitDate,
		customer.Gender,
		customer.CustomerType,
		customer.CustomerAgeRange,
		customer.PurchasedCategories,
		customer.Description,
		customer.MarketerNote,
		customer.SignatureURL,
		customer.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) SoftDelete(ctx context.Context, id shared.ID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE customers SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) ListBirthdayToday(ctx context.Context, day time.Time) ([]domain.BirthdayRecipient, error) {
	const q = `
SELECT id, first_name, last_name, phone
FROM customers
WHERE deleted_at IS NULL
  AND birthday IS NOT NULL
  AND EXTRACT(MONTH FROM birthday) = EXTRACT(MONTH FROM $1::date)
  AND EXTRACT(DAY FROM birthday) = EXTRACT(DAY FROM $1::date)
`
	rows, err := r.pool.Query(ctx, q, day)
	if err != nil {
		return nil, fmt.Errorf("list birthday today: %w", err)
	}
	defer rows.Close()

	var out []domain.BirthdayRecipient
	for rows.Next() {
		var rec domain.BirthdayRecipient
		if err := rows.Scan(&rec.ID, &rec.FirstName, &rec.LastName, &rec.Phone); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repository) ListPhonesByFilter(ctx context.Context, filter domain.ListFilter) ([]domain.BirthdayRecipient, error) {
	q := buildListQuery(filter, nil)
	sql := `
SELECT c.id, c.first_name, c.last_name, c.phone
` + listBaseFrom + `
` + q.where + `
ORDER BY c.created_at DESC`
	rows, err := r.pool.Query(ctx, sql, q.args...)
	if err != nil {
		return nil, fmt.Errorf("list phones by filter: %w", err)
	}
	defer rows.Close()

	var out []domain.BirthdayRecipient
	for rows.Next() {
		var rec domain.BirthdayRecipient
		if err := rows.Scan(&rec.ID, &rec.FirstName, &rec.LastName, &rec.Phone); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repository) WasBirthdaySMSSent(ctx context.Context, customerID shared.ID, day time.Time) (bool, error) {
	const q = `SELECT EXISTS(
		SELECT 1 FROM sms_birthday_log WHERE customer_id = $1 AND sent_on = $2::date
	)`
	var exists bool
	err := r.pool.QueryRow(ctx, q, customerID, day).Scan(&exists)
	return exists, err
}

func (r *Repository) RecordBirthdaySMS(ctx context.Context, customerID shared.ID, day time.Time, status, errMsg string) error {
	const q = `
INSERT INTO sms_birthday_log (customer_id, sent_on, status, error_message)
VALUES ($1, $2::date, $3, NULLIF($4, ''))
ON CONFLICT (customer_id, sent_on) DO NOTHING
`
	_, err := r.pool.Exec(ctx, q, customerID, day, status, errMsg)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanCustomer(row scannable) (*domain.Customer, error) {
	var c domain.Customer
	err := row.Scan(
		&c.ID,
		&c.FirstName,
		&c.LastName,
		&c.Job,
		&c.Phone,
		&c.Email,
		&c.Address,
		&c.MelliCode,
		&c.PostalCode,
		&c.Birthday,
		&c.MarriageDate,
		&c.ImportantDate,
		&c.FirstVisitDate,
		&c.Gender,
		&c.CustomerType,
		&c.CustomerAgeRange,
		&c.PurchasedCategories,
		&c.Description,
		&c.MarketerNote,
		&c.SignatureURL,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if c.CustomerType == "" {
		c.CustomerType = domain.DefaultCustomerType
	}
	c.PurchasedCategories = domain.StringSliceOrEmpty(c.PurchasedCategories)
	return &c, nil
}
