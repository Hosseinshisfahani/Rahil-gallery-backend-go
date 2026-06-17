package seed

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/security"
)

const markerEmail = "admin@rehil.gallery"

// Dev credentials printed after a successful seed run.
const (
	AdminEmail       = markerEmail
	AdminPassword    = "Admin1234"
	StaffEmail       = "staff@rehil.gallery"
	StaffPassword    = "Staff1234"
	CustomerPassword = "Customer12"
)

var ErrAlreadySeeded = errors.New("database already seeded; use --reset or --force")

type roleIDs struct {
	customer string
	admin    string
	staff    string
}

type Runner struct {
	pool *pgxpool.Pool
	hash security.BcryptHasher
	opts Options
}

func Run(ctx context.Context, pool *pgxpool.Pool, opts Options) error {
	opts.Customers = opts.normalizedCustomers()
	opts.Products = opts.normalizedProducts()
	r := &Runner{pool: pool, hash: security.NewBcryptHasher(), opts: opts}

	if opts.Reset {
		log.Println("seed: clearing dev seed data...")
		if err := r.reset(ctx); err != nil {
			return fmt.Errorf("reset: %w", err)
		}
	}

	seeded, err := r.isSeeded(ctx)
	if err != nil {
		return err
	}
	if seeded && !opts.Force && !opts.Reset {
		return ErrAlreadySeeded
	}

	roles, err := r.loadRoles(ctx)
	if err != nil {
		return err
	}

	log.Printf("seed: target %d customers (%d fixtures + %d bulk)",
		opts.Customers, MinCustomers, opts.bulkCustomerCount())
	log.Printf("seed: target %d products (%d fixtures + %d bulk)",
		opts.Products, MinProducts, opts.bulkProductCount())

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.seedStaff(ctx, tx, roles); err != nil {
		return fmt.Errorf("staff: %w", err)
	}
	if err := r.seedCatalog(ctx, tx); err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	if err := r.seedFixtureCustomers(ctx, tx, roles); err != nil {
		return fmt.Errorf("fixtures: %w", err)
	}
	if err := r.seedFixtureCommerce(ctx, tx); err != nil {
		return fmt.Errorf("fixture commerce: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if err := r.refreshCommerceStats(ctx); err != nil {
		return fmt.Errorf("fixture commerce stats: %w", err)
	}

	if err := r.seedBulkCustomers(ctx, roles.customer); err != nil {
		return fmt.Errorf("bulk customers: %w", err)
	}
	if err := r.seedBulkCommerce(ctx); err != nil {
		return fmt.Errorf("bulk commerce: %w", err)
	}
	if err := r.seedBulkProducts(ctx); err != nil {
		return fmt.Errorf("bulk products: %w", err)
	}

	if err := r.backfillMissingCRMProfiles(ctx); err != nil {
		return fmt.Errorf("backfill CRM profiles: %w", err)
	}

	log.Println("seed: done")
	log.Printf("  customers total: %d", opts.Customers)
	log.Printf("  products total:  %d", opts.Products)
	log.Printf("  admin  %s / %s", AdminEmail, AdminPassword)
	log.Printf("  staff  %s / %s", StaffEmail, StaffPassword)
	log.Printf("  sample customer (password set)  customer@rehil.gallery / %s", CustomerPassword)
	log.Printf("  bulk customers: phones %sXXXXXXXX, emails seed-*@rehil.dev", bulkPhonePrefix)
	log.Printf("  bulk products:  SKUs %sXXXXXXXX", bulkProductSKUPrefix)

	return nil
}

func (r *Runner) isSeeded(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`,
		markerEmail,
	).Scan(&exists)
	return exists, err
}

func (r *Runner) loadRoles(ctx context.Context) (roleIDs, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM roles`)
	if err != nil {
		return roleIDs{}, err
	}
	defer rows.Close()

	var roles roleIDs
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return roleIDs{}, err
		}
		switch name {
		case "customer":
			roles.customer = id
		case "admin":
			roles.admin = id
		case "staff":
			roles.staff = id
		}
	}
	if roles.customer == "" || roles.admin == "" || roles.staff == "" {
		return roleIDs{}, errors.New("required roles not found; run migrations first")
	}
	return roles, rows.Err()
}

func (r *Runner) reset(ctx context.Context) error {
	if err := r.resetBulk(ctx); err != nil {
		return err
	}
	if err := r.resetBulkProducts(ctx); err != nil {
		return err
	}
	return r.resetFixtures(ctx)
}

func (r *Runner) resetBulk(ctx context.Context) error {
	log.Println("seed: clearing bulk data (phone prefix +98900...)...")

	stmts := []string{
		`DELETE FROM customer_audit_log WHERE target_user_id IN (SELECT id FROM users WHERE phone LIKE $1)`,
		`DELETE FROM customer_notes WHERE user_id IN (SELECT id FROM users WHERE phone LIKE $1)`,
		`DELETE FROM wishlist_items WHERE user_id IN (SELECT id FROM users WHERE phone LIKE $1)`,
		`DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE order_number LIKE 'RG-BULK-%')`,
		`DELETE FROM payments WHERE order_id IN (SELECT id FROM orders WHERE order_number LIKE 'RG-BULK-%')`,
		`DELETE FROM order_status_history WHERE order_id IN (SELECT id FROM orders WHERE order_number LIKE 'RG-BULK-%')`,
		`DELETE FROM orders WHERE order_number LIKE 'RG-BULK-%'`,
		`DELETE FROM customer_profiles WHERE user_id IN (SELECT id FROM users WHERE phone LIKE $1)`,
		`DELETE FROM users WHERE phone LIKE $1`,
	}

	pattern := bulkPhonePrefix + "%"
	for _, q := range stmts {
		var err error
		if strings.Contains(q, "$1") {
			_, err = r.pool.Exec(ctx, q, pattern)
		} else {
			_, err = r.pool.Exec(ctx, q)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) resetFixtures(ctx context.Context) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cIDs := customerIDs()
	sIDs := append(staffIDs(), cIDs...)
	oIDs := orderIDs()
	vIDs := variantIDs()
	pIDs := productIDs()
	itemIDs := []uuid.UUID{IDOrderItemSara1, IDOrderItemSara2, IDOrderItemAli1, IDOrderItemNeda1}

	imageIDs := []uuid.UUID{IDImageSolitaire, IDImageEternity, IDImagePearl}

	queries := []struct {
		q    string
		args []any
	}{
		{`DELETE FROM customer_audit_log WHERE target_user_id = ANY($1) OR admin_id = ANY($2)`, []any{cIDs, sIDs}},
		{`DELETE FROM customer_notes WHERE user_id = ANY($1)`, []any{cIDs}},
		{`DELETE FROM wishlist_items WHERE user_id = ANY($1)`, []any{cIDs}},
		{`DELETE FROM order_status_history WHERE order_id = ANY($1)`, []any{oIDs}},
		{`DELETE FROM order_items WHERE id = ANY($1)`, []any{itemIDs}},
		{`DELETE FROM payments WHERE order_id = ANY($1)`, []any{oIDs}},
		{`DELETE FROM orders WHERE id = ANY($1)`, []any{oIDs}},
		{`DELETE FROM product_collections WHERE product_id = ANY($1)`, []any{pIDs}},
		{`DELETE FROM product_images WHERE id = ANY($1)`, []any{imageIDs}},
		{`DELETE FROM inventory_items WHERE variant_id = ANY($1)`, []any{vIDs}},
		{`DELETE FROM product_variants WHERE id = ANY($1)`, []any{vIDs}},
		{`DELETE FROM products WHERE id = ANY($1)`, []any{pIDs}},
		{`DELETE FROM collections WHERE id = ANY($1)`, []any{fixtureCollectionIDs()}},
		{`DELETE FROM categories WHERE id = ANY($1)`, []any{fixtureCategoryIDs()}},
		{`DELETE FROM customer_profiles WHERE user_id = ANY($1)`, []any{cIDs}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{sIDs}},
	}

	for _, item := range queries {
		if _, err := tx.Exec(ctx, item.q, item.args...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func exec(ctx context.Context, tx pgx.Tx, query string, args ...any) error {
	_, err := tx.Exec(ctx, query, args...)
	return err
}
