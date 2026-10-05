package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
)

const maxLastErrorRunes = 2000

type JobRepository struct {
	pool *pgxpool.Pool
}

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{pool: pool}
}

type CreateJobInput struct {
	CreatedBy           *model.ID
	Message             string
	FilterSnapshot      any
	Matched             int
	SkippedInvalidPhone int
	BatchCount          int
}

func (r *JobRepository) Create(ctx context.Context, in CreateJobInput) (uuid.UUID, error) {
	snap, err := json.Marshal(in.FilterSnapshot)
	if err != nil {
		return uuid.Nil, err
	}
	id := uuid.New()
	const q = `
INSERT INTO sms_jobs (
  id, created_by, message, filter_snapshot, status,
  matched_count, skipped_invalid_phone, batch_count
) VALUES ($1,$2,$3,$4,'pending',$5,$6,$7)`
	_, err = r.pool.Exec(ctx, q, id, in.CreatedBy, in.Message, snap, in.Matched, in.SkippedInvalidPhone, in.BatchCount)
	return id, err
}

func (r *JobRepository) MarkRunning(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE sms_jobs SET status='running', updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *JobRepository) MarkFinished(ctx context.Context, id uuid.UUID, status string, sent, failed int, lastError string) error {
	lastError = truncateRunes(strings.TrimSpace(lastError), maxLastErrorRunes)
	var errPtr *string
	if lastError != "" {
		errPtr = &lastError
	}
	_, err := r.pool.Exec(ctx, `
UPDATE sms_jobs
SET status=$2, sent_count=$3, failed_count=$4, last_error=$5, completed_at=NOW(), updated_at=NOW()
WHERE id=$1`, id, status, sent, failed, errPtr)
	return err
}

type Job struct {
	ID                  uuid.UUID  `json:"id"`
	Status              string     `json:"status"`
	Message             string     `json:"message"`
	MatchedCount        int        `json:"matched"`
	SkippedInvalidPhone int        `json:"skippedInvalidPhone"`
	SentCount           int        `json:"sent"`
	FailedCount         int        `json:"failed"`
	BatchCount          int        `json:"batches"`
	SellerNote          *string    `json:"sellerNote"`
	LastError           *string    `json:"lastError"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	CompletedAt         *time.Time `json:"completedAt"`
}

type ListResult struct {
	Items      []Job
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

func (r *JobRepository) List(ctx context.Context, page, perPage int) (ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sms_jobs`).Scan(&total); err != nil {
		return ListResult{}, err
	}

	offset := (page - 1) * perPage
	rows, err := r.pool.Query(ctx, `
SELECT id, status, message,
       matched_count, skipped_invalid_phone, sent_count, failed_count, batch_count,
       seller_note, last_error, created_at, updated_at, completed_at
FROM sms_jobs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()

	items := make([]Job, 0)
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.ID, &j.Status, &j.Message,
			&j.MatchedCount, &j.SkippedInvalidPhone, &j.SentCount, &j.FailedCount, &j.BatchCount,
			&j.SellerNote, &j.LastError, &j.CreatedAt, &j.UpdatedAt, &j.CompletedAt,
		); err != nil {
			return ListResult{}, err
		}
		items = append(items, j)
	}

	totalPages := 1
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return ListResult{Items: items, Total: total, Page: page, PerPage: perPage, TotalPages: totalPages}, rows.Err()
}

func (r *JobRepository) UpdateSellerNote(ctx context.Context, id uuid.UUID, note string) (*Job, error) {
	const q = `
UPDATE sms_jobs
SET seller_note = NULLIF($2, ''), updated_at = NOW()
WHERE id = $1
RETURNING id, status, message,
          matched_count, skipped_invalid_phone, sent_count, failed_count, batch_count,
          seller_note, last_error, created_at, updated_at, completed_at`
	var j Job
	err := r.pool.QueryRow(ctx, q, id, strings.TrimSpace(note)).Scan(
		&j.ID, &j.Status, &j.Message,
		&j.MatchedCount, &j.SkippedInvalidPhone, &j.SentCount, &j.FailedCount, &j.BatchCount,
		&j.SellerNote, &j.LastError, &j.CreatedAt, &j.UpdatedAt, &j.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}
