package identity

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) FindByID(ctx context.Context, id shared.ID) (*identity.Role, error) {
	const q = `SELECT id, name, description, created_at FROM roles WHERE id = $1`

	var role identity.Role
	var description *string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&role.ID, &role.Name, &description, &role.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	role.Description = derefString(description)
	return &role, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*identity.Role, error) {
	const q = `SELECT id, name, description, created_at FROM roles WHERE name = $1`

	var role identity.Role
	var description *string
	err := r.pool.QueryRow(ctx, q, name).Scan(
		&role.ID, &role.Name, &description, &role.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
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

var _ identity.RoleRepository = (*RoleRepository)(nil)
