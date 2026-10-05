package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
)

var ErrPhoneAlreadyExists = errors.New("phone number already registered")

type CustomerService struct {
	customer   *repository.Repository
	signatures Store
}

func NewCustomerService(customer *repository.Repository, signaturesDir string) *CustomerService {
	return &CustomerService{
		customer:   customer,
		signatures: Store{Dir: signaturesDir},
	}
}

func (s *CustomerService) Create(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	if err := s.prepare(ctx, customer, nil); err != nil {
		return nil, err
	}
	now := time.Now()
	customer.ID = uuid.New()
	customer.CreatedAt = now
	customer.UpdatedAt = now
	return s.customer.Create(ctx, customer)
}

func (s *CustomerService) Update(ctx context.Context, id model.ID, customer *model.Customer) (*model.Customer, error) {
	existing, err := s.customer.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.prepare(ctx, customer, &id); err != nil {
		return nil, err
	}
	customer.ID = id
	customer.CreatedAt = existing.CreatedAt
	customer.UpdatedAt = time.Now()
	customer.SignatureURL = existing.SignatureURL
	return s.customer.Update(ctx, customer)
}

func (s *CustomerService) Delete(ctx context.Context, id model.ID) error {
	if err := s.signatures.Delete(id); err != nil {
		return err
	}
	return s.customer.SoftDelete(ctx, id)
}

func (s *CustomerService) UploadSignature(
	ctx context.Context,
	id model.ID,
	contentType string,
	data []byte,
) (*model.Customer, error) {
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
	return s.customer.Update(ctx, customer)
}

func (s *CustomerService) DeleteSignature(ctx context.Context, id model.ID) (*model.Customer, error) {
	customer, err := s.customer.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.signatures.Delete(id); err != nil {
		return nil, err
	}

	customer.SignatureURL = nil
	customer.UpdatedAt = time.Now()
	return s.customer.Update(ctx, customer)
}

func (s *CustomerService) prepare(ctx context.Context, customer *model.Customer, excludeID *model.ID) error {
	model.NormalizeCustomer(customer)
	customer.FirstName = strings.TrimSpace(customer.FirstName)
	customer.LastName = strings.TrimSpace(customer.LastName)
	customer.Phone = strings.TrimSpace(customer.Phone)
	customer.Job = trimOptional(customer.Job)
	customer.Email = trimOptional(customer.Email)
	customer.Address = trimOptional(customer.Address)
	customer.Gender = trimOptional(customer.Gender)
	customer.CustomerAgeRange = trimOptional(customer.CustomerAgeRange)
	customer.Description = trimOptional(customer.Description)
	customer.MarketerNote = trimOptional(customer.MarketerNote)
	customer.SignatureURL = trimOptional(customer.SignatureURL)

	if customer.Phone == "" || customer.FirstName == "" || customer.LastName == "" {
		return model.ErrInvalidInput
	}

	exists, err := s.customer.PhoneExists(ctx, customer.Phone, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrPhoneAlreadyExists
	}

	if customer.Email != nil {
		email := normalizeCustomerEmail(*customer.Email)
		exists, err := s.customer.EmailExists(ctx, email, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return ErrEmailAlreadyExists
		}
		customer.Email = &email
	}
	return nil
}

func normalizeCustomerEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func trimOptional(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
