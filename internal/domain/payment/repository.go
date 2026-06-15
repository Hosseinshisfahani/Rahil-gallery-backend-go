package payment

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type PaymentRepository interface {
	FindByID(ctx context.Context, id shared.ID) (*Payment, error)
	FindByOrderID(ctx context.Context, orderID shared.ID) ([]Payment, error)
	Create(ctx context.Context, payment *Payment) error
	Update(ctx context.Context, payment *Payment) error
}
