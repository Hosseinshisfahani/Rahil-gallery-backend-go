package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByID(ctx context.Context, id model.ID) (*model.User, error) {
	const q = `
		SELECT id, role_id, email, phone, password_hash, first_name, last_name,
		       status, email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	return r.scanOne(ctx, q, id)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `
		SELECT id, role_id, email, phone, password_hash, first_name, last_name,
		       status, email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	return r.scanOne(ctx, q, email)
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
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

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
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

func (r *UserRepository) scanOne(ctx context.Context, query string, arg any) (*model.User, error) {
	row := r.pool.QueryRow(ctx, query, arg)

	var u model.User
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
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	u.Email = email
	u.Phone = phone
	u.PasswordHash = passwordHash
	u.Status = model.UserStatus(status)
	u.EmailVerifiedAt = emailVerifiedAt
	u.LastLoginAt = lastLoginAt
	u.DeletedAt = deletedAt

	return &u, nil
}

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) FindByID(ctx context.Context, id model.ID) (*model.Role, error) {
	const q = `SELECT id, name, description, created_at FROM roles WHERE id = $1`

	var role model.Role
	var description *string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&role.ID, &role.Name, &description, &role.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	role.Description = derefString(description)
	return &role, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	const q = `SELECT id, name, description, created_at FROM roles WHERE name = $1`

	var role model.Role
	var description *string
	err := r.pool.QueryRow(ctx, q, name).Scan(
		&role.ID, &role.Name, &description, &role.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	role.Description = derefString(description)
	return &role, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, q,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.RevokedAt, token.CreatedAt,
	)
	return err
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	var t model.RefreshToken
	var revokedAt *time.Time
	err := r.pool.QueryRow(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &revokedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	t.RevokedAt = revokedAt
	return &t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id model.ID) error {
	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
