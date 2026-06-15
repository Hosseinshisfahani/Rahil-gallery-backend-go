package identity

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByID(ctx context.Context, id shared.ID) (*identity.User, error) {
	const q = `
		SELECT id, role_id, email, phone, password_hash, first_name, last_name,
		       status, email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	return r.scanOne(ctx, q, id)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	const q = `
		SELECT id, role_id, email, phone, password_hash, first_name, last_name,
		       status, email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	return r.scanOne(ctx, q, email)
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*identity.User, error) {
	const q = `
		SELECT id, role_id, email, phone, password_hash, first_name, last_name,
		       status, email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL`

	return r.scanOne(ctx, q, phone)
}

func (r *UserRepository) Create(ctx context.Context, user *identity.User) error {
	const q = `
		INSERT INTO users (
			id, role_id, email, phone, password_hash, first_name, last_name,
			status, email_verified_at, last_login_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	_, err := r.pool.Exec(ctx, q,
		user.ID, user.RoleID, user.Email, user.Phone, user.PasswordHash,
		user.FirstName, user.LastName, user.Status,
		user.EmailVerifiedAt, user.LastLoginAt, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *identity.User) error {
	const q = `
		UPDATE users SET
			role_id = $2, email = $3, phone = $4, password_hash = $5,
			first_name = $6, last_name = $7, status = $8,
			email_verified_at = $9, last_login_at = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, q,
		user.ID, user.RoleID, user.Email, user.Phone, user.PasswordHash,
		user.FirstName, user.LastName, user.Status,
		user.EmailVerifiedAt, user.LastLoginAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) SoftDelete(ctx context.Context, id shared.ID) error {
	const q = `UPDATE users SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	now := time.Now()
	tag, err := r.pool.Exec(ctx, q, id, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *UserRepository) scanOne(ctx context.Context, query string, arg any) (*identity.User, error) {
	row := r.pool.QueryRow(ctx, query, arg)

	var u identity.User
	var status string
	var phone, email, passwordHash *string
	var emailVerifiedAt, lastLoginAt, deletedAt *time.Time

	err := row.Scan(
		&u.ID, &u.RoleID, &email, &phone, &passwordHash,
		&u.FirstName, &u.LastName, &status,
		&emailVerifiedAt, &lastLoginAt, &u.CreatedAt, &u.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	u.Email = email
	u.Phone = phone
	u.PasswordHash = passwordHash
	u.Status = identity.UserStatus(status)
	u.EmailVerifiedAt = emailVerifiedAt
	u.LastLoginAt = lastLoginAt
	u.DeletedAt = deletedAt

	return &u, nil
}

var _ identity.UserRepository = (*UserRepository)(nil)
