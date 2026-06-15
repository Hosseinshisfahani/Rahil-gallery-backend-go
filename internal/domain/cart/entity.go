package cart

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type CartStatus string

const (
	CartStatusActive    CartStatus = "active"
	CartStatusMerged    CartStatus = "merged"
	CartStatusAbandoned CartStatus = "abandoned"
	CartStatusConverted CartStatus = "converted"
)

type Cart struct {
	ID         shared.ID
	UserID     *shared.ID
	GuestToken *string
	Status     CartStatus
	ExpiresAt  *time.Time
	Items      []CartItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CartItem struct {
	ID                shared.ID
	CartID            shared.ID
	VariantID         shared.ID
	Quantity          int
	UnitPriceSnapshot shared.Money
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
