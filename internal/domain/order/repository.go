package order

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type OrderRepository interface {
	FindByID(ctx context.Context, id shared.ID) (*Order, error)
	FindByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	ListByUserID(ctx context.Context, userID shared.ID, limit, offset int) ([]Order, error)
	Create(ctx context.Context, order *Order) error
	UpdateStatus(ctx context.Context, id shared.ID, status OrderStatus) error
	AppendStatusHistory(ctx context.Context, entry *OrderStatusHistory) error
}
