package model

import (
	"time"

	"github.com/google/uuid"
)

type ID = uuid.UUID

const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
	RoleStaff    = "staff"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

type Role struct {
	ID          ID
	Name        string
	Description string
	CreatedAt   time.Time
}

type User struct {
	ID              ID
	RoleID          ID
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
	ID        ID
	UserID    ID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
