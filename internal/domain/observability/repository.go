package observability

import (
	"context"
	"time"
)

type Repository interface {
	Insert(ctx context.Context, input IngestInput) (*Event, error)
	List(ctx context.Context, filter ListFilter) ([]Event, int, error)
	Summary(ctx context.Context, since time.Time) (*Summary, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}
