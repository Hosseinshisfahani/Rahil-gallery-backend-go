package inventory

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type InventoryItem struct {
	ID                shared.ID
	VariantID         shared.ID
	Quantity          int
	ReservedQuantity  int
	LowStockThreshold int
	UpdatedAt         time.Time
}

func (i InventoryItem) Available() int {
	return i.Quantity - i.ReservedQuantity
}

func (i InventoryItem) IsLowStock() bool {
	return i.Available() <= i.LowStockThreshold
}
