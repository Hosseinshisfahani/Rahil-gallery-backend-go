package promotion

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "percentage"
	DiscountTypeFixedAmount DiscountType = "fixed_amount"
)

type Coupon struct {
	ID             shared.ID
	Code           string
	DiscountType   DiscountType
	DiscountValue  float64
	MinOrderAmount float64
	MaxUses        *int
	UsedCount      int
	ValidFrom      time.Time
	ValidUntil     time.Time
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CouponRedemption struct {
	ID             shared.ID
	CouponID       shared.ID
	OrderID        shared.ID
	UserID         shared.ID
	DiscountAmount shared.Money
	RedeemedAt     time.Time
}
