package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/observability"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Insert(ctx context.Context, input domain.IngestInput) (*domain.Event, error) {
	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	const q = `
INSERT INTO observability_events (
  source, level, message, stack_trace, route, method, status_code, request_id, user_agent, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id, source, level, message, stack_trace, route, method, status_code, request_id, user_agent, metadata, created_at`

	row := r.pool.QueryRow(ctx, q,
		input.Source, input.Level, input.Message,
		input.StackTrace, input.Route, input.Method, input.StatusCode,
		input.RequestID, input.UserAgent, metadata,
	)
	return scanEvent(row)
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Event, int, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 25
	}
	if perPage > 100 {
		perPage = 100
	}

	where, args := buildListWhere(filter)
	offset := (page - 1) * perPage

	countQ := "SELECT COUNT(*) FROM observability_events" + where
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := `
SELECT id, source, level, message, stack_trace, route, method, status_code, request_id, user_agent, metadata, created_at
FROM observability_events` + where + `
ORDER BY created_at DESC
LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)

	args = append(args, perPage, offset)
	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Event, 0, perPage)
	for rows.Next() {
		ev, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *ev)
	}
	return out, total, rows.Err()
}

func (r *Repository) Summary(ctx context.Context, since time.Time) (*domain.Summary, error) {
	const totalQ = `SELECT COUNT(*) FROM observability_events WHERE created_at >= $1`
	const last24Q = `SELECT COUNT(*) FROM observability_events WHERE created_at >= now() - interval '24 hours'`

	summary := &domain.Summary{
		RetentionDays: domain.DefaultRetentionDays,
		BySource:      map[string]int{},
		Daily:         []domain.DailyCount{},
	}

	if err := r.pool.QueryRow(ctx, totalQ, since).Scan(&summary.Total); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, last24Q).Scan(&summary.Last24Hours); err != nil {
		return nil, err
	}

	bySourceRows, err := r.pool.Query(ctx, `
SELECT source, COUNT(*) FROM observability_events
WHERE created_at >= $1
GROUP BY source`, since)
	if err != nil {
		return nil, err
	}
	defer bySourceRows.Close()
	for bySourceRows.Next() {
		var source string
		var count int
		if err := bySourceRows.Scan(&source, &count); err != nil {
			return nil, err
		}
		summary.BySource[source] = count
	}

	dailyRows, err := r.pool.Query(ctx, `
SELECT to_char(date_trunc('day', created_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS day,
       source, level, COUNT(*)
FROM observability_events
WHERE created_at >= $1
GROUP BY 1, source, level
ORDER BY 1 DESC`, since)
	if err != nil {
		return nil, err
	}
	defer dailyRows.Close()
	for dailyRows.Next() {
		var row domain.DailyCount
		if err := dailyRows.Scan(&row.Day, &row.Source, &row.Level, &row.Count); err != nil {
			return nil, err
		}
		summary.Daily = append(summary.Daily, row)
	}

	return summary, nil
}

func (r *Repository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM observability_events WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func buildListWhere(filter domain.ListFilter) (string, []any) {
	parts := []string{"WHERE created_at >= $1"}
	args := []any{filter.From}
	if filter.From == nil {
		parts[0] = "WHERE created_at >= now() - interval '7 days'"
		args = nil
	}
	idx := len(args) + 1

	add := func(cond string, val any) {
		parts = append(parts, fmt.Sprintf(cond, idx))
		args = append(args, val)
		idx++
	}

	if filter.Source != "" {
		add("source = $%d", filter.Source)
	}
	if filter.Level != "" {
		add("level = $%d", filter.Level)
	}
	if filter.Route != "" {
		add("route ILIKE $%d", "%"+filter.Route+"%")
	}
	if filter.To != nil {
		add("created_at <= $%d", *filter.To)
	}

	return " " + strings.Join(parts, " AND "), args
}

type scannable interface {
	Scan(dest ...any) error
}

func scanEvent(row scannable) (*domain.Event, error) {
	var ev domain.Event
	if err := row.Scan(
		&ev.ID, &ev.Source, &ev.Level, &ev.Message,
		&ev.StackTrace, &ev.Route, &ev.Method, &ev.StatusCode,
		&ev.RequestID, &ev.UserAgent, &ev.Metadata, &ev.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &ev, nil
}
