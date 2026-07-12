package customer

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Repository interface {
	List(ctx context.Context, filter ListFilter, page, perPage int) (ListResult, error)
	Get(ctx context.Context, id shared.ID) (*Customer, error)
	PhoneExists(ctx context.Context, phone string, excludeID *shared.ID) (bool, error)
	EmailExists(ctx context.Context, email string, excludeID *shared.ID) (bool, error)
	Create(ctx context.Context, customer *Customer) error
	Update(ctx context.Context, customer *Customer) error
	SoftDelete(ctx context.Context, id shared.ID) error
}
