package customer

import (
	"context"
	"time"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

func (s *Service) UploadSignature(
	ctx context.Context,
	id shared.ID,
	contentType string,
	data []byte,
) (*domain.Customer, error) {
	customer, err := s.customer.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	url, err := s.signatures.Save(id, contentType, data)
	if err != nil {
		return nil, err
	}

	customer.SignatureURL = &url
	customer.UpdatedAt = time.Now()

	if err := s.customer.Update(ctx, customer); err != nil {
		return nil, err
	}

	return s.customer.Get(ctx, id)
}

func (s *Service) DeleteSignature(ctx context.Context, id shared.ID) (*domain.Customer, error) {
	customer, err := s.customer.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.signatures.Delete(id); err != nil {
		return nil, err
	}

	customer.SignatureURL = nil
	customer.UpdatedAt = time.Now()

	if err := s.customer.Update(ctx, customer); err != nil {
		return nil, err
	}

	return s.customer.Get(ctx, id)
}
