package identity

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

type AddressType string

const (
	AddressTypeShipping AddressType = "shipping"
	AddressTypeBilling  AddressType = "billing"
	AddressTypeBoth     AddressType = "both"
)

type Role struct {
	ID          shared.ID
	Name        string
	Description string
	CreatedAt   time.Time
}

type User struct {
	ID              shared.ID
	RoleID          shared.ID
	Email           *string
	Phone           *string
	PasswordHash    *string
	FirstName       string
	LastName        string
	Status          UserStatus
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type RefreshToken struct {
	ID        shared.ID
	UserID    shared.ID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type UserAddress struct {
	ID             shared.ID
	UserID         shared.ID
	AddressType    AddressType
	Label          *string
	RecipientName  string
	Phone          string
	Province       string
	City           string
	PostalCode     string
	AddressLine    string
	IsDefault      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
