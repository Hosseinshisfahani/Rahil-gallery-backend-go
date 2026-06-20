package observability

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/observability"
)

var (
	ErrInvalidSource = errors.New("invalid observability source")
	ErrInvalidLevel  = errors.New("invalid observability level")
	ErrEmptyMessage  = errors.New("message is required")
)

type Service struct {
	repo           domain.Repository
	retentionDays  int
	startOnce      sync.Once
}

func NewService(repo domain.Repository, retentionDays int) *Service {
	if retentionDays < 1 {
		retentionDays = domain.DefaultRetentionDays
	}
	return &Service{repo: repo, retentionDays: retentionDays}
}

func (s *Service) StartBackgroundJobs(ctx context.Context) {
	s.startOnce.Do(func() {
		go s.retentionLoop(ctx)
	})
}

func (s *Service) Record(ctx context.Context, input domain.IngestInput) error {
	if err := validateInput(input); err != nil {
		return err
	}
	_, err := s.repo.Insert(ctx, input)
	return err
}

func (s *Service) RecordAPIError(ctx context.Context, message string, route, method, requestID, userAgent string, statusCode int, stack string) {
	input := domain.IngestInput{
		Source:     domain.SourceAPI,
		Level:      domain.LevelError,
		Message:    truncate(message, 4000),
		Route:      optionalString(route),
		Method:     optionalString(method),
		StatusCode: &statusCode,
		RequestID:  optionalString(requestID),
		UserAgent:  optionalString(truncate(userAgent, 512)),
	}
	if stack != "" {
		input.StackTrace = optionalString(truncate(stack, 8000))
	}
	if err := s.Record(ctx, input); err != nil {
		log.Printf("observability: record api error: %v", err)
	}
}

func (s *Service) Ingest(ctx context.Context, input domain.IngestInput) (*domain.Event, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	return s.repo.Insert(ctx, input)
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]domain.Event, int, error) {
	if filter.From == nil {
		since := s.retentionCutoff()
		filter.From = &since
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) Summary(ctx context.Context) (*domain.Summary, error) {
	summary, err := s.repo.Summary(ctx, s.retentionCutoff())
	if err != nil {
		return nil, err
	}
	summary.RetentionDays = s.retentionDays
	return summary, nil
}

func (s *Service) retentionCutoff() time.Time {
	return time.Now().UTC().AddDate(0, 0, -s.retentionDays)
}

func (s *Service) retentionLoop(ctx context.Context) {
	s.runRetention(ctx)
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runRetention(ctx)
		}
	}
}

func (s *Service) runRetention(ctx context.Context) {
	cutoff := s.retentionCutoff()
	n, err := s.repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		log.Printf("observability: retention cleanup failed: %v", err)
		return
	}
	if n > 0 {
		log.Printf("observability: deleted %d events older than %s", n, cutoff.Format(time.RFC3339))
	}
}

func validateInput(input domain.IngestInput) error {
	input.Source = strings.TrimSpace(input.Source)
	input.Level = strings.TrimSpace(input.Level)
	input.Message = strings.TrimSpace(input.Message)

	switch input.Source {
	case domain.SourceAPI, domain.SourceClient:
	default:
		return ErrInvalidSource
	}
	switch input.Level {
	case domain.LevelError, domain.LevelWarn, domain.LevelInfo:
	default:
		return ErrInvalidLevel
	}
	if input.Message == "" {
		return ErrEmptyMessage
	}
	if len(input.Metadata) > 0 && !json.Valid(input.Metadata) {
		return errors.New("metadata must be valid JSON")
	}
	return nil
}

func optionalString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
