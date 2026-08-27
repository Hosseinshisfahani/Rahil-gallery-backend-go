package sms

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	domainsms "github.com/rahil-gallery/rahil-gallery-server/internal/domain/sms"
	smspg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/sms"
)

const maxRecipients = 20000

var (
	ErrEmptyMessage   = errors.New("message is required")
	ErrTooManyTargets = errors.New("too many recipients")
	ErrNoTargets      = errors.New("no valid recipients")
)

type BulkService struct {
	Customers customer.Repository
	Jobs      *smspg.JobRepository
	SMS       domainsms.Provider
	Sender    string
	BatchSize int
	Workers   int
}

type BulkAcceptResult struct {
	JobID               uuid.UUID `json:"jobId"`
	Accepted            bool      `json:"accepted"`
	Matched             int       `json:"matched"`
	SkippedInvalidPhone int       `json:"skippedInvalidPhone"`
	Batches             int       `json:"batches"`
}

func (s *BulkService) Accept(ctx context.Context, createdBy *shared.ID, message string, filter customer.ListFilter) (*BulkAcceptResult, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, ErrEmptyMessage
	}
	if len([]rune(message)) > 900 {
		return nil, ErrEmptyMessage
	}

	recs, err := s.Customers.ListPhonesByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(recs) > maxRecipients {
		return nil, ErrTooManyTargets
	}

	receptors := make([]string, 0, len(recs))
	skipped := 0
	for _, rec := range recs {
		n, err := domainsms.NormalizeReceptor(rec.Phone)
		if err != nil {
			skipped++
			continue
		}
		receptors = append(receptors, n)
	}
	if len(receptors) == 0 {
		return nil, ErrNoTargets
	}

	batchSize := s.BatchSize
	if batchSize < 1 || batchSize > 200 {
		batchSize = 200
	}
	batches := (len(receptors) + batchSize - 1) / batchSize

	jobID, err := s.Jobs.Create(ctx, smspg.CreateJobInput{
		CreatedBy:           createdBy,
		Message:             message,
		FilterSnapshot:      filter,
		Matched:             len(recs),
		SkippedInvalidPhone: skipped,
		BatchCount:          batches,
	})
	if err != nil {
		return nil, err
	}

	go s.runJob(jobID, receptors, message, batchSize)

	return &BulkAcceptResult{
		JobID:               jobID,
		Accepted:            true,
		Matched:             len(recs),
		SkippedInvalidPhone: skipped,
		Batches:             batches,
	}, nil
}

func (s *BulkService) runJob(jobID uuid.UUID, receptors []string, message string, batchSize int) {
	ctx := context.Background()
	_ = s.Jobs.MarkRunning(ctx, jobID)

	workers := s.Workers
	if workers < 1 {
		workers = 1
	}

	type chunk struct{ items []string }
	chunks := make(chan chunk, workers*2)
	var sent int64
	var failed int64
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ch := range chunks {
				_, err := s.SMS.SendBulk(ctx, s.Sender, ch.items, message)
				if err != nil {
					atomic.AddInt64(&failed, int64(len(ch.items)))
					log.Printf("sms bulk job=%s batch failed: %v", jobID, err)
					continue
				}
				atomic.AddInt64(&sent, int64(len(ch.items)))
			}
		}()
	}

	for i := 0; i < len(receptors); i += batchSize {
		end := i + batchSize
		if end > len(receptors) {
			end = len(receptors)
		}
		chunks <- chunk{items: receptors[i:end]}
	}
	close(chunks)
	wg.Wait()

	status := "completed"
	if failed > 0 && sent > 0 {
		status = "completed_with_errors"
	} else if failed > 0 {
		status = "failed"
	}
	_ = s.Jobs.MarkFinished(ctx, jobID, status, int(sent), int(failed))
}
