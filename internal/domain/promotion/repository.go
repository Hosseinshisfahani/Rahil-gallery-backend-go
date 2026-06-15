package promotion

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type CouponRepository interface {
	FindByCode(ctx context.Context, code string) (*Coupon, error)
	IncrementUsedCount(ctx context.Context, id shared.ID) error
	CreateRedemption(ctx context.Context, redemption *CouponRedemption) error
}
