package customer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

func (r *Repository) ListSavedViews(ctx context.Context, ownerID shared.ID, viewType *domain.ListViewType) ([]domain.SavedListView, error) {
	q := `
SELECT id, owner_id, name, view_type::text, filters, is_shared, position, created_at, updated_at
FROM admin_customer_list_views
WHERE owner_id = $1 OR is_shared = TRUE`
	args := []any{ownerID}
	if viewType != nil {
		q += ` AND view_type = $2::customer_list_view_type`
		args = append(args, string(*viewType))
	}
	q += ` ORDER BY position ASC, name ASC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSavedViews(rows)
}

func (r *Repository) GetSavedView(ctx context.Context, id, requesterID shared.ID) (*domain.SavedListView, error) {
	const q = `
SELECT id, owner_id, name, view_type::text, filters, is_shared, position, created_at, updated_at
FROM admin_customer_list_views
WHERE id = $1 AND (owner_id = $2 OR is_shared = TRUE)`

	rows, err := r.pool.Query(ctx, q, id, requesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views, err := scanSavedViews(rows)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, shared.ErrNotFound
	}
	return &views[0], nil
}

func (r *Repository) CreateSavedView(ctx context.Context, view *domain.SavedListView) error {
	filtersJSON, err := json.Marshal(view.Filters)
	if err != nil {
		return err
	}

	const q = `
INSERT INTO admin_customer_list_views (id, owner_id, name, view_type, filters, is_shared, position, created_at, updated_at)
VALUES ($1, $2, $3, $4::customer_list_view_type, $5, $6, $7, $8, $9)`

	_, err = r.pool.Exec(ctx, q,
		view.ID, view.OwnerID, view.Name, string(view.ViewType), filtersJSON,
		view.IsShared, view.Position, view.CreatedAt, view.UpdatedAt,
	)
	return mapSavedViewErr(err)
}

func (r *Repository) UpdateSavedView(ctx context.Context, view *domain.SavedListView) error {
	filtersJSON, err := json.Marshal(view.Filters)
	if err != nil {
		return err
	}

	const q = `
UPDATE admin_customer_list_views
SET name = $3, view_type = $4::customer_list_view_type, filters = $5,
    is_shared = $6, position = $7, updated_at = $8
WHERE id = $1 AND owner_id = $2`

	tag, err := r.pool.Exec(ctx, q,
		view.ID, view.OwnerID, view.Name, string(view.ViewType), filtersJSON,
		view.IsShared, view.Position, view.UpdatedAt,
	)
	if err != nil {
		return mapSavedViewErr(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteSavedView(ctx context.Context, id, ownerID shared.ID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM admin_customer_list_views WHERE id = $1 AND owner_id = $2`,
		id, ownerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *Repository) CountCustomersBySegment(ctx context.Context) ([]domain.SegmentSummary, error) {
	const q = `
SELECT cp.segment, COUNT(*)::int
FROM customer_profiles cp
INNER JOIN users u ON u.id = cp.user_id AND u.deleted_at IS NULL
INNER JOIN roles r ON r.id = u.role_id AND r.name = $1
GROUP BY cp.segment`

	rows, err := r.pool.Query(ctx, q, identity.RoleCustomer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var seg string
		var count int
		if err := rows.Scan(&seg, &count); err != nil {
			return nil, err
		}
		counts[seg] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	order := []domain.Segment{
		domain.SegmentVIP,
		domain.SegmentNew,
		domain.SegmentActive,
		domain.SegmentReturning,
		domain.SegmentInactive,
	}
	out := make([]domain.SegmentSummary, 0, len(order))
	for _, seg := range order {
		label := domain.SegmentLabels[seg]
		if label == "" {
			label = string(seg)
		}
		out = append(out, domain.SegmentSummary{
			Segment: seg,
			Label:   label,
			Count:   counts[string(seg)],
		})
	}
	return out, nil
}

func scanSavedViews(rows pgx.Rows) ([]domain.SavedListView, error) {
	var views []domain.SavedListView
	for rows.Next() {
		var v domain.SavedListView
		var viewType string
		var filtersJSON []byte
		if err := rows.Scan(
			&v.ID, &v.OwnerID, &v.Name, &viewType, &filtersJSON,
			&v.IsShared, &v.Position, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		v.ViewType = domain.ListViewType(viewType)
		if len(filtersJSON) > 0 {
			if err := json.Unmarshal(filtersJSON, &v.Filters); err != nil {
				return nil, err
			}
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

func mapSavedViewErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return shared.ErrConflict
	}
	return err
}
