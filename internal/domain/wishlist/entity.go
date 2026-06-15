package wishlist

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type WishlistItem struct {
	UserID    shared.ID
	VariantID shared.ID
	CreatedAt time.Time
}
