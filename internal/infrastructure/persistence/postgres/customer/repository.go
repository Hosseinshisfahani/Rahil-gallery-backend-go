package customer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
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

	q := buildListQuery(filter, []any{identity.RoleCustomer})

	if filter.SkipCount {
		return r.listWithoutCount(ctx, q, page, perPage)
	}

	var total int
	var items []domain.ListRow
	var countErr, listErr error

	if shouldParallelList(filter) {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			total, countErr = r.countCustomers(ctx, q)
		}()
		go func() {
			defer wg.Done()
			items, listErr = r.fetchCustomerPage(ctx, q, page, perPage, (page-1)*perPage)
		}()
		wg.Wait()
		if countErr != nil {
			return domain.ListResult{}, countErr
		}
		if listErr != nil {
			return domain.ListResult{}, listErr
		}
	} else {
		total, countErr = r.countCustomers(ctx, q)
		if countErr != nil {
			return domain.ListResult{}, countErr
		}

		totalPages := int(math.Max(1, math.Ceil(float64(total)/float64(perPage))))
		if page > totalPages {
			page = totalPages
		}
		offset := (page - 1) * perPage

		items, listErr = r.fetchCustomerPage(ctx, q, page, perPage, offset)
		if listErr != nil {
			return domain.ListResult{}, listErr
		}
	}

	totalPages := int(math.Max(1, math.Ceil(float64(total)/float64(perPage))))
	if page > totalPages && total > 0 {
		page = totalPages
		items, listErr = r.fetchCustomerPage(ctx, q, page, perPage, (page-1)*perPage)
		if listErr != nil {
			return domain.ListResult{}, listErr
		}
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
	items, err := r.fetchCustomerPage(ctx, q, page, perPage+1, offset)
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

func shouldParallelList(filter domain.ListFilter) bool {
	if filter.HasAdvancedFilters() {
		return true
	}
	return strings.TrimSpace(filter.QuickSearch) != ""
}

func (r *Repository) countCustomers(ctx context.Context, q listQuery) (int, error) {
	countQ := fmt.Sprintf(`%s SELECT COUNT(*) %s %s`, q.cte, q.from, q.where)
	var total int
	if err := r.pool.QueryRow(ctx, countQ, q.args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) fetchCustomerPage(ctx context.Context, q listQuery, page, perPage, offset int) ([]domain.ListRow, error) {
	if offset < 0 {
		offset = 0
	}
	listQ := fmt.Sprintf(`%s
SELECT
  u.id,
  TRIM(u.first_name || ' ' || u.last_name) AS full_name,
  COALESCE(u.phone, '') AS phone,
  u.email,
  u.created_at,
  cp.last_activity_at,
  cp.last_purchase_at,
  COALESCE(cp.total_orders, 0),
  COALESCE(cp.total_ltv, 0)::float8,
  %s AS segment,
  %s AS status,
  COALESCE(cp.is_vip, FALSE),
  COALESCE(cp.tags, '{}'),
  ` + crmCustomerTypeExpr + `,
  ` + crmPurchasedCategoriesExpr + `,
  ` + crmAgeRangeExpr + `,
  ` + crmGenderExpr + `
%s %s
ORDER BY u.created_at DESC
LIMIT $%d OFFSET $%d`, q.cte, segmentExpr, statusExpr, q.from, q.where, len(q.args)+1, len(q.args)+2)

	args := append(append([]any{}, q.args...), perPage, offset)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ListRow, 0, perPage)
	for rows.Next() {
		var row domain.ListRow
		var email *string
		var lastActivity, lastPurchase *time.Time
		var tags []string
		var customerType *string
		var purchasedCategories []string
		var customerAgeRange, gender *string
		if err := rows.Scan(
			&row.ID, &row.FullName, &row.Phone, &email,
			&row.RegisteredAt, &lastActivity, &lastPurchase,
			&row.TotalOrders, &row.TotalLTV,
			&row.Segment, &row.Status, &row.IsVIP, &tags,
			&customerType, &purchasedCategories,
			&customerAgeRange, &gender,
		); err != nil {
			return nil, err
		}
		row.Email = email
		row.LastActivityAt = lastActivity
		row.LastPurchaseDate = lastPurchase
		row.Tags = tags
		row.CustomerType = customerType
		row.PurchasedCategories = purchasedCategories
		row.CustomerAgeRange = customerAgeRange
		row.Gender = gender
		domain.NormalizeListRowCRM(&row)
		items = append(items, row)
	}
	return items, rows.Err()
}

func (r *Repository) GetDetail(ctx context.Context, userID shared.ID) (*domain.Detail, error) {
	const q = `
SELECT
  u.id,
  TRIM(u.first_name || ' ' || u.last_name),
  COALESCE(u.phone, ''),
  u.email,
  u.created_at,
  cp.last_activity_at,
  cp.last_purchase_at,
  COALESCE(cp.total_orders, 0),
  COALESCE(cp.total_ltv, 0)::float8,
  ` + segmentExpr + `,
  ` + statusExpr + `,
  COALESCE(cp.is_vip, FALSE),
  COALESCE(cp.tags, '{}'),
  ` + crmCustomerTypeExpr + `,
  ` + crmPurchasedCategoriesExpr + `,
  COALESCE(cp.locale, 'fa'),
  cp.default_ring_size,
  cp.first_purchase_at,
  cp.vip_source::text,
  cp.block_reason,
  cp.block_note,
  cp.import_mode::text,
  cp.import_profile
FROM users u
INNER JOIN roles rl ON rl.id = u.role_id AND rl.name = $2
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
WHERE u.id = $1 AND u.deleted_at IS NULL`

	var d domain.Detail
	var email *string
	var vipSource, blockReason, importMode *string
	var blockNote *string
	var defaultRing *string
	var lastActivity, lastPurchase, firstPurchase *time.Time
	var customerType *string
	var purchasedCategories []string
	var importProfile []byte

	err := r.pool.QueryRow(ctx, q, userID, identity.RoleCustomer).Scan(
		&d.ID, &d.FullName, &d.Phone, &email,
		&d.RegisteredAt, &lastActivity, &lastPurchase,
		&d.TotalOrders, &d.TotalLTV,
		&d.Segment, &d.Status, &d.IsVIP, &d.Tags,
		&customerType, &purchasedCategories,
		&d.Locale, &defaultRing, &firstPurchase,
		&vipSource, &blockReason, &blockNote, &importMode, &importProfile,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	d.Email = email
	d.LastActivityAt = lastActivity
	d.LastPurchaseDate = lastPurchase
	d.FirstPurchaseDate = firstPurchase
	d.DefaultRingSize = defaultRing
	d.BlockNote = blockNote
	d.CustomerType = customerType
	d.PurchasedCategories = purchasedCategories
	d.ImportProfile = importProfile
	domain.NormalizeListRowCRM(&d.ListRow)

	if vipSource != nil {
		v := domain.VIPSource(*vipSource)
		d.VIPSource = &v
	}
	if blockReason != nil {
		br := domain.BlockReason(*blockReason)
		d.BlockReason = &br
	}
	if importMode != nil {
		im := domain.ImportMode(*importMode)
		d.ImportMode = &im
	}

	if d.TotalOrders > 0 {
		d.AverageOrderValue = d.TotalLTV / float64(d.TotalOrders)
	}
	if d.TotalOrders > 1 {
		d.RepeatPurchaseRate = float64(d.TotalOrders-1) / float64(d.TotalOrders)
	}
	d.PurchaseFrequency = float64(d.TotalOrders)
	d.EngagementScore = computeEngagementScore(d.TotalOrders, d.WishlistCount)
	d.FunnelPosition = deriveFunnelPosition(d.TotalOrders, d.WishlistCount)

	orders, err := r.ListOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	d.Orders = orders

	wishlist, err := r.ListWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}
	d.Wishlist = wishlist
	d.WishlistCount = len(wishlist)

	notes, err := r.ListNotes(ctx, userID)
	if err != nil {
		return nil, err
	}
	d.Notes = notes

	audit, err := r.ListAudit(ctx, userID)
	if err != nil {
		return nil, err
	}
	d.AuditLog = audit

	return &d, nil
}

func computeEngagementScore(orders, wishlist int) int {
	score := orders*15 + wishlist*5
	if score > 100 {
		return 100
	}
	return score
}

func deriveFunnelPosition(orders, wishlist int) string {
	switch {
	case orders > 0:
		return "purchased"
	case wishlist > 0:
		return "consideration"
	default:
		return "awareness"
	}
}

func (r *Repository) PhoneExists(ctx context.Context, phone string, excludeUserID *shared.ID) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND deleted_at IS NULL`
	args := []any{phone}
	if excludeUserID != nil {
		q += ` AND id <> $2`
		args = append(args, *excludeUserID)
	}
	q += `)`

	var exists bool
	err := r.pool.QueryRow(ctx, q, args...).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertProfile(ctx context.Context, profile *domain.Profile) error {
	const q = `
INSERT INTO customer_profiles (
  user_id, locale, default_ring_size, is_vip, vip_source, import_mode, import_profile,
  tags, block_reason, block_note, last_activity_at, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (user_id) DO UPDATE SET
  locale = EXCLUDED.locale,
  default_ring_size = EXCLUDED.default_ring_size,
  is_vip = EXCLUDED.is_vip,
  vip_source = EXCLUDED.vip_source,
  import_mode = EXCLUDED.import_mode,
  import_profile = EXCLUDED.import_profile,
  tags = EXCLUDED.tags,
  block_reason = EXCLUDED.block_reason,
  block_note = EXCLUDED.block_note,
  last_activity_at = EXCLUDED.last_activity_at,
  updated_at = EXCLUDED.updated_at`

	var vipSource, importMode, blockReason *string
	if profile.VIPSource != nil {
		s := string(*profile.VIPSource)
		vipSource = &s
	}
	if profile.ImportMode != nil {
		s := string(*profile.ImportMode)
		importMode = &s
	}
	if profile.BlockReason != nil {
		s := string(*profile.BlockReason)
		blockReason = &s
	}

	_, err := r.pool.Exec(ctx, q,
		profile.UserID, profile.Locale, profile.DefaultRingSize, profile.IsVIP,
		vipSource, importMode, profile.ImportProfile, profile.Tags,
		blockReason, profile.BlockNote, profile.LastActivityAt,
		profile.CreatedAt, profile.UpdatedAt,
	)
	return err
}

func (r *Repository) GetProfile(ctx context.Context, userID shared.ID) (*domain.Profile, error) {
	const q = `
SELECT user_id, locale, default_ring_size, is_vip, vip_source::text, import_mode::text,
       import_profile, tags, block_reason, block_note, last_activity_at, created_at, updated_at
FROM customer_profiles WHERE user_id = $1`

	var p domain.Profile
	var vipSource, importMode, blockReason *string
	var importProfile []byte

	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.Locale, &p.DefaultRingSize, &p.IsVIP,
		&vipSource, &importMode, &importProfile, &p.Tags,
		&blockReason, &p.BlockNote, &p.LastActivityAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	p.ImportProfile = importProfile
	if vipSource != nil {
		v := domain.VIPSource(*vipSource)
		p.VIPSource = &v
	}
	if importMode != nil {
		im := domain.ImportMode(*importMode)
		p.ImportMode = &im
	}
	if blockReason != nil {
		br := domain.BlockReason(*blockReason)
		p.BlockReason = &br
	}

	return &p, nil
}

func (r *Repository) AddNote(ctx context.Context, note *domain.Note) error {
	const q = `INSERT INTO customer_notes (id, user_id, author_id, body, created_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, q, note.ID, note.UserID, note.AuthorID, note.Body, note.CreatedAt)
	return err
}

func (r *Repository) ListNotes(ctx context.Context, userID shared.ID) ([]domain.Note, error) {
	const q = `
SELECT n.id, n.user_id, n.author_id,
       TRIM(a.first_name || ' ' || a.last_name) AS author_name,
       n.body, n.created_at
FROM customer_notes n
INNER JOIN users a ON a.id = n.author_id
WHERE n.user_id = $1
ORDER BY n.created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []domain.Note
	for rows.Next() {
		var n domain.Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.AuthorID, &n.Author, &n.Body, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *Repository) AppendAudit(ctx context.Context, entry *domain.AuditEntry) error {
	const q = `
INSERT INTO customer_audit_log (id, admin_id, target_user_id, action, reason, details, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, q,
		entry.ID, entry.AdminID, entry.TargetUserID, string(entry.Action),
		entry.Reason, entry.Details, entry.CreatedAt,
	)
	return err
}

func (r *Repository) ListAudit(ctx context.Context, userID shared.ID) ([]domain.AuditEntry, error) {
	const q = `
SELECT a.id, a.admin_id, TRIM(u.first_name || ' ' || u.last_name),
       a.target_user_id, a.action, a.reason, a.details, a.created_at
FROM customer_audit_log a
INNER JOIN users u ON u.id = a.admin_id
WHERE a.target_user_id = $1
ORDER BY a.created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		var action string
		if err := rows.Scan(
			&e.ID, &e.AdminID, &e.AdminName, &e.TargetUserID,
			&action, &e.Reason, &e.Details, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		e.Action = domain.AuditAction(action)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *Repository) ListOrders(ctx context.Context, userID shared.ID) ([]domain.OrderSummary, error) {
	const q = `
SELECT o.id, o.placed_at, o.total_amount::float8, o.status::text,
       COALESCE((SELECT SUM(oi.quantity) FROM order_items oi WHERE oi.order_id = o.id), 0)::int,
       EXISTS(SELECT 1 FROM order_status_history h WHERE h.order_id = o.id AND h.to_status = 'refunded')
FROM orders o
WHERE o.user_id = $1
ORDER BY o.placed_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.OrderSummary
	for rows.Next() {
		var o domain.OrderSummary
		if err := rows.Scan(&o.ID, &o.Date, &o.Total, &o.Status, &o.ItemCount, &o.HasReturn); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *Repository) ListWishlist(ctx context.Context, userID shared.ID) ([]domain.WishlistSummary, error) {
	const q = `
SELECT pv.id, p.name, COALESCE(c.name, 'Jewelry'), (p.base_price + pv.price_adjustment)::float8,
       w.created_at, FALSE, NULL::text
FROM wishlist_items w
INNER JOIN product_variants pv ON pv.id = w.variant_id
INNER JOIN products p ON p.id = pv.product_id
LEFT JOIN categories c ON c.id = p.category_id
WHERE w.user_id = $1
ORDER BY w.created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.WishlistSummary
	for rows.Next() {
		var w domain.WishlistSummary
		if err := rows.Scan(
			&w.ID, &w.ProductName, &w.Category, &w.Price,
			&w.SavedAt, &w.IsConfiguration, &w.ConfigurationSummary,
		); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

var _ domain.Repository = (*Repository)(nil)
