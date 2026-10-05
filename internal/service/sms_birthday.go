package service

import (
	"context"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
)

type BirthdayRunner struct {
	Customers *repository.Repository
	SMS       model.Provider
	Template  string
}

func (r *BirthdayRunner) Run(ctx context.Context, now time.Time) error {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	recs, err := r.Customers.ListBirthdayToday(ctx, day)
	if err != nil {
		return err
	}

	var sent, skippedInvalid, skippedDup, failed int
	for _, rec := range recs {
		already, err := r.Customers.WasBirthdaySMSSent(ctx, rec.ID, day)
		if err != nil {
			return err
		}
		if already {
			skippedDup++
			continue
		}

		receptor, err := model.NormalizeReceptor(rec.Phone)
		if err != nil {
			skippedInvalid++
			_ = r.Customers.RecordBirthdaySMS(ctx, rec.ID, day, BirthdayStatusSkippedInvalidPhone, err.Error())
			continue
		}

		token := lookupToken(rec.FirstName)
		err = r.SMS.SendLookup(ctx, receptor, r.Template, map[string]string{"token": token})
		if err != nil {
			failed++
			_ = r.Customers.RecordBirthdaySMS(ctx, rec.ID, day, BirthdayStatusFailed, err.Error())
			continue
		}

		sent++
		_ = r.Customers.RecordBirthdaySMS(ctx, rec.ID, day, BirthdayStatusSent, "")
	}

	log.Printf("birthday_sms_run candidates=%d sent=%d skipped_invalid=%d skipped_dup=%d failed=%d",
		len(recs), sent, skippedInvalid, skippedDup, failed)
	return nil
}

// %token cannot contain spaces (Kavenegar rule for your approved template).
func lookupToken(firstName string) string {
	name := strings.TrimSpace(firstName)
	if name == "" {
		return LookupFallbackToken
	}
	// take first whitespace-separated word, strip remaining spaces just in case
	fields := strings.Fields(name)
	token := fields[0]
	var b strings.Builder
	for _, r := range token {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return LookupFallbackToken
	}
	return out
}
