package review

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type ProductReviewRepository interface {
	ListApprovedByProductID(ctx context.Context, productID shared.ID, limit, offset int) ([]ProductReview, error)
	Create(ctx context.Context, review *ProductReview) error
}
