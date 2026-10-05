package seed

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
)

var ErrProductionBootstrapped = errors.New("production bootstrap already applied; use --update-password to rotate admin credentials")

// ProductionOptions configures the minimal production database bootstrap.
type ProductionOptions struct {
	AdminEmail     string
	AdminPassword  string
	AdminFirstName string
	AdminLastName  string
	AdminPhone     string

	StaffEmail     string
	StaffPassword  string
	StaffFirstName string
	StaffLastName  string
	StaffPhone     string

	UpdatePassword bool
}

func (o ProductionOptions) includesStaff() bool {
	return strings.TrimSpace(o.StaffEmail) != "" || strings.TrimSpace(o.StaffPassword) != ""
}

func (o ProductionOptions) normalized() (ProductionOptions, error) {
	out := o
	out.AdminEmail = normalizeEmail(o.AdminEmail)
	out.StaffEmail = normalizeEmail(o.StaffEmail)
	out.AdminFirstName = strings.TrimSpace(o.AdminFirstName)
	out.AdminLastName = strings.TrimSpace(o.AdminLastName)
	out.AdminPhone = strings.TrimSpace(o.AdminPhone)
	out.StaffFirstName = strings.TrimSpace(o.StaffFirstName)
	out.StaffLastName = strings.TrimSpace(o.StaffLastName)
	out.StaffPhone = strings.TrimSpace(o.StaffPhone)

	if out.AdminFirstName == "" {
		out.AdminFirstName = "Admin"
	}
	if out.AdminLastName == "" {
		out.AdminLastName = "Rehil"
	}
	if out.StaffFirstName == "" {
		out.StaffFirstName = "Staff"
	}
	if out.StaffLastName == "" {
		out.StaffLastName = "User"
	}

	if err := validateProductionCredentials(out.AdminEmail, out.AdminPassword, out.AdminFirstName, out.AdminLastName); err != nil {
		return ProductionOptions{}, fmt.Errorf("admin: %w", err)
	}

	if out.includesStaff() {
		if out.StaffEmail == "" || out.StaffPassword == "" {
			return ProductionOptions{}, errors.New("set both SEED_STAFF_EMAIL and SEED_STAFF_PASSWORD to create a staff account")
		}
		if err := validateProductionCredentials(out.StaffEmail, out.StaffPassword, out.StaffFirstName, out.StaffLastName); err != nil {
			return ProductionOptions{}, fmt.Errorf("staff: %w", err)
		}
	}

	return out, nil
}

func validateProductionCredentials(email, password, firstName, lastName string) error {
	if normalizeEmail(email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return errors.New("first and last name are required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !passwordStrongEnough(password) {
		return errors.New("password must contain at least one letter and one digit")
	}
	return nil
}

func passwordStrongEnough(password string) bool {
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// RunProduction seeds only production essentials (admin account, optional staff).
// Roles are created by migrations; no demo customers, catalog, or bulk data.
func RunProduction(ctx context.Context, pool *pgxpool.Pool, opts ProductionOptions) error {
	opts, err := opts.normalized()
	if err != nil {
		return err
	}

	r := &Runner{pool: pool, hash: service.NewBcryptHasher()}

	roles, err := r.loadRoles(ctx)
	if err != nil {
		return err
	}

	if opts.UpdatePassword {
		if err := r.updateProductionPasswords(ctx, roles, opts); err != nil {
			return err
		}
		log.Println("seed-prod: admin password updated")
		log.Printf("  admin  %s", opts.AdminEmail)
		return nil
	}

	hasAdmin, err := r.hasAdminAccount(ctx)
	if err != nil {
		return err
	}
	if hasAdmin {
		return ErrProductionBootstrapped
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.insertProductionUser(ctx, tx, roles.admin, productionUser{
		email:     opts.AdminEmail,
		password:  opts.AdminPassword,
		firstName: opts.AdminFirstName,
		lastName:  opts.AdminLastName,
		phone:     opts.AdminPhone,
	}); err != nil {
		return fmt.Errorf("admin: %w", err)
	}

	if opts.includesStaff() {
		if err := r.insertProductionUser(ctx, tx, roles.staff, productionUser{
			email:     opts.StaffEmail,
			password:  opts.StaffPassword,
			firstName: opts.StaffFirstName,
			lastName:  opts.StaffLastName,
			phone:     opts.StaffPhone,
		}); err != nil {
			return fmt.Errorf("staff: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Println("seed-prod: production bootstrap complete")
	log.Printf("  admin  %s", opts.AdminEmail)
	if opts.includesStaff() {
		log.Printf("  staff  %s", opts.StaffEmail)
	}
	log.Println("  (passwords were not printed — store them securely)")

	return nil
}

type productionUser struct {
	email, password, firstName, lastName, phone string
}

func (r *Runner) hasAdminAccount(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM users u
  INNER JOIN roles r ON r.id = u.role_id AND r.name = 'admin'
  WHERE u.deleted_at IS NULL
)`).Scan(&exists)
	return exists, err
}

func (r *Runner) insertProductionUser(ctx context.Context, tx pgx.Tx, roleID string, user productionUser) error {
	hash, err := r.hash.Hash(user.password)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	id := uuid.New().String()

	var phone any
	if user.phone != "" {
		phone = user.phone
	}

	return exec(ctx, tx, `
INSERT INTO users (
	id, role_id, email, phone, password_hash, first_name, last_name,
	status, email_verified_at, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,$8,$8)`,
		id, roleID, user.email, phone, hash, user.firstName, user.lastName, now,
	)
}

func (r *Runner) updateProductionPasswords(ctx context.Context, roles roleIDs, opts ProductionOptions) error {
	hash, err := r.hash.Hash(opts.AdminPassword)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx, `
UPDATE users u
SET password_hash = $1, updated_at = NOW()
FROM roles r
WHERE u.role_id = r.id
  AND r.name = 'admin'
  AND u.email = $2
  AND u.deleted_at IS NULL`,
		hash, opts.AdminEmail,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no active admin user found with email %q", opts.AdminEmail)
	}

	if !opts.includesStaff() {
		return nil
	}

	staffHash, err := r.hash.Hash(opts.StaffPassword)
	if err != nil {
		return err
	}

	tag, err = r.pool.Exec(ctx, `
UPDATE users u
SET password_hash = $1, updated_at = NOW()
FROM roles r
WHERE u.role_id = r.id
  AND r.name = 'staff'
  AND u.email = $2
  AND u.deleted_at IS NULL`,
		staffHash, opts.StaffEmail,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no active staff user found with email %q", opts.StaffEmail)
	}

	return nil
}
