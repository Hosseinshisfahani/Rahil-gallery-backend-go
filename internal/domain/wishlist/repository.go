package wishlist

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type WishlistRepository interface {
	ListByUserID(ctx context.Context, userID shared.ID) ([]WishlistItem, error)
	Add(ctx context.Context, item *WishlistItem) error
	Remove(ctx context.Context, userID, variantID shared.ID) error
}
