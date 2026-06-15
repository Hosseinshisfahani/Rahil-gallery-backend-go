package inventory

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type InventoryRepository interface {
	FindByVariantID(ctx context.Context, variantID shared.ID) (*InventoryItem, error)
	Reserve(ctx context.Context, variantID shared.ID, quantity int) error
	Release(ctx context.Context, variantID shared.ID, quantity int) error
	CommitReservation(ctx context.Context, variantID shared.ID, quantity int) error
}
