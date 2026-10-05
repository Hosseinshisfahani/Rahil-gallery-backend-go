package seed

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
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
	hash service.BcryptHasher
	opts Options
}

func Run(ctx context.Context, pool *pgxpool.Pool, opts Options) error {
	opts.Customers = opts.normalizedCustomers()
	r := &Runner{pool: pool, hash: service.NewBcryptHasher(), opts: opts}

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

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.seedStaff(ctx, tx, roles); err != nil {
		return fmt.Errorf("staff: %w", err)
	}
	if err := r.seedFixtureCustomers(ctx, tx, roles); err != nil {
		return fmt.Errorf("fixtures: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if err := r.seedBulkCustomers(ctx, roles.customer); err != nil {
		return fmt.Errorf("bulk customers: %w", err)
	}

	log.Println("seed: done")
	log.Printf("  customers total: %d", opts.Customers)
	log.Printf("  admin  %s / %s", AdminEmail, AdminPassword)
	log.Printf("  staff  %s / %s", StaffEmail, StaffPassword)
	log.Printf("  sample customer (password set)  customer@rehil.gallery / %s", CustomerPassword)
	log.Printf("  bulk customers: phones %sXXXXXXXX, emails seed-*@rehil.dev", bulkPhonePrefix)

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
	return r.resetFixtures(ctx)
}

func (r *Runner) resetBulk(ctx context.Context) error {
	log.Println("seed: clearing bulk data (phone prefix +98900...)...")

	stmts := []string{
		`DELETE FROM users WHERE id IN (SELECT id FROM customers WHERE phone LIKE $1)`,
		`DELETE FROM customers WHERE phone LIKE $1`,
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

	queries := []struct {
		q    string
		args []any
	}{
		{`DELETE FROM users WHERE id = ANY($1)`, []any{cIDs}},
		{`DELETE FROM customers WHERE id = ANY($1)`, []any{cIDs}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{staffIDs()}},
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
