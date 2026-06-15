package cart

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type CartRepository interface {
	FindByID(ctx context.Context, id shared.ID) (*Cart, error)
	FindActiveByUserID(ctx context.Context, userID shared.ID) (*Cart, error)
	FindActiveByGuestToken(ctx context.Context, guestToken string) (*Cart, error)
	Create(ctx context.Context, cart *Cart) error
	Update(ctx context.Context, cart *Cart) error
	UpsertItem(ctx context.Context, item *CartItem) error
	RemoveItem(ctx context.Context, cartID, itemID shared.ID) error
}
