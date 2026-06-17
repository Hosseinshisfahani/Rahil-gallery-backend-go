package customer

import (
	"context"
	"encoding/json"
	"time"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

func (s *Service) UploadSignature(
	ctx context.Context,
	adminID, userID shared.ID,
	contentType string,
	data []byte,
) (*domain.Detail, error) {
	if _, err := s.customer.GetDetail(ctx, userID); err != nil {
		return nil, err
	}

	url, err := s.signatures.Save(userID, contentType, data)
	if err != nil {
		return nil, err
	}

	if err := s.setProfileSignature(ctx, userID, url); err != nil {
		return nil, err
	}

	details := "Signature uploaded"
	_ = s.audit(ctx, adminID, userID, domain.AuditActionProfileEdit, nil, &details)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) DeleteSignature(ctx context.Context, adminID, userID shared.ID) (*domain.Detail, error) {
	if _, err := s.customer.GetDetail(ctx, userID); err != nil {
		return nil, err
	}

	if err := s.signatures.Delete(userID); err != nil {
		return nil, err
	}

	if err := s.setProfileSignature(ctx, userID, ""); err != nil {
		return nil, err
	}

	details := "Signature removed"
	_ = s.audit(ctx, adminID, userID, domain.AuditActionProfileEdit, nil, &details)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) setProfileSignature(ctx context.Context, userID shared.ID, signatureURL string) error {
	profile, err := s.customer.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	profile.ImportProfile, err = mergeImportProfileSignature(profile.ImportProfile, signatureURL)
	if err != nil {
		return err
	}
	profile.UpdatedAt = time.Now()

	return s.customer.UpsertProfile(ctx, profile)
}

func mergeImportProfileSignature(raw json.RawMessage, signatureURL string) (json.RawMessage, error) {
	m := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
	}

	if signatureURL == "" {
		delete(m, "signature")
	} else {
		m["signature"] = signatureURL
	}

	return json.Marshal(m)
}
