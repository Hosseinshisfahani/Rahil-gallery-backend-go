package review

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type ProductReview struct {
	ID         shared.ID
	ProductID  shared.ID
	UserID     shared.ID
	Rating     int16
	Title      *string
	Body       *string
	IsApproved bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
