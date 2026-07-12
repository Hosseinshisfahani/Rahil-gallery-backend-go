package customer

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/storage/customersignature"
)

var ErrPhoneAlreadyExists = errors.New("phone number already registered")

type Service struct {
	customer   domain.Repository
	signatures customersignature.Store
}

func NewService(customer domain.Repository, signaturesDir string) *Service {
	return &Service{
		customer:   customer,
		signatures: customersignature.Store{Dir: signaturesDir},
	}
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter, page, perPage int) (domain.ListResult, error) {
	return s.customer.List(ctx, filter, page, perPage)
}

func (s *Service) Get(ctx context.Context, id shared.ID) (*domain.Customer, error) {
	return s.customer.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, in domain.Input) (*domain.Customer, error) {
	domain.NormalizeInput(&in)

	phone := strings.TrimSpace(in.Phone)
	first := strings.TrimSpace(in.FirstName)
	last := strings.TrimSpace(in.LastName)
	
	if phone == "" || first == "" || last == "" {
		return nil, shared.ErrInvalidInput
	}

	exists, err := s.customer.PhoneExists(ctx, phone, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneAlreadyExists
	}

	if in.Email != nil && strings.TrimSpace(*in.Email) != "" {
		email := normalizeEmail(*in.Email)
		exists, err := s.customer.EmailExists(ctx, email, nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, shared.ErrConflict
		}
		in.Email = &email
	}

	now := time.Now()
	customer := inputToCustomer(uuid.New(), in, now, now)

	if err := s.customer.Create(ctx, customer); err != nil {
		return nil, err
	}

	return s.customer.Get(ctx, customer.ID)
}

func (s *Service) Update(ctx context.Context, id shared.ID, in domain.Input) (*domain.Customer, error) {
	domain.NormalizeInput(&in)

	existing, err := s.customer.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	phone := strings.TrimSpace(in.Phone)
	first := strings.TrimSpace(in.FirstName)
	last := strings.TrimSpace(in.LastName)
	if phone == "" || first == "" || last == "" {
		return nil, shared.ErrInvalidInput
	}

	exists, err := s.customer.PhoneExists(ctx, phone, &id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneAlreadyExists
	}

	if in.Email != nil && strings.TrimSpace(*in.Email) != "" {
		email := normalizeEmail(*in.Email)
		exists, err := s.customer.EmailExists(ctx, email, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, shared.ErrConflict
		}
		in.Email = &email
	}

	updated := inputToCustomer(id, in, existing.CreatedAt, time.Now())
	updated.SignatureURL = existing.SignatureURL

	if err := s.customer.Update(ctx, updated); err != nil {
		return nil, err
	}

	return s.customer.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id shared.ID) error {
	if err := s.signatures.Delete(id); err != nil {
		return err
	}
	return s.customer.SoftDelete(ctx, id)
}

func inputToCustomer(id shared.ID, in domain.Input, createdAt, updatedAt time.Time) *domain.Customer {
	return &domain.Customer{
		ID:                  id,
		FirstName:           strings.TrimSpace(in.FirstName),
		LastName:            strings.TrimSpace(in.LastName),
		Job:                 trimOptional(in.Job),
		Phone:               strings.TrimSpace(in.Phone),
		Email:               trimOptional(in.Email),
		Address:             trimOptional(in.Address),
		Birthday:            in.Birthday,
		MarriageDate:        in.MarriageDate,
		ImportantDate:       in.ImportantDate,
		FirstVisitDate:      in.FirstVisitDate,
		Gender:              trimOptional(in.Gender),
		CustomerType:        in.CustomerType,
		CustomerAgeRange:    trimOptional(in.CustomerAgeRange),
		PurchasedCategories: domain.StringSliceOrEmpty(in.PurchasedCategories),
		Description:         trimOptional(in.Description),
		SignatureURL:        trimOptional(in.SignatureURL),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}
}

func normalizeEmail(email string) string {
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
