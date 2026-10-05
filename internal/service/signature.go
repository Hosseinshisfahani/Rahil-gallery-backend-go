package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidType = errors.New("signature must be PNG, JPEG, or WebP")
	ErrTooLarge    = errors.New("signature file exceeds 2MB limit")
)

var allowedTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

type Store struct {
	Dir string
}

func (s *Store) Save(customerID uuid.UUID, contentType string, data []byte) (string, error) {
	if s.Dir == "" {
		return "", fmt.Errorf("customer signatures directory is not configured")
	}

	ext, ok := allowedTypes[strings.ToLower(strings.TrimSpace(contentType))]
	if !ok {
		return "", ErrInvalidType
	}
	if len(data) == 0 {
		return "", ErrInvalidType
	}
	if len(data) > MaxSignatureBytes {
		return "", ErrTooLarge
	}

	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return "", err
	}

	s.removeExisting(customerID)

	filename := customerID.String() + ext
	path := filepath.Join(s.Dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}

	return "/static/customer-signatures/" + filename, nil
}

func (s *Store) Delete(customerID uuid.UUID) error {
	if s.Dir == "" {
		return nil
	}
	s.removeExisting(customerID)
	return nil
}

func (s *Store) removeExisting(customerID uuid.UUID) {
	for _, ext := range allowedTypes {
		_ = os.Remove(filepath.Join(s.Dir, customerID.String()+ext))
	}
}
