package customer

import (
	"context"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type BirthdayRecipient struct {
	ID        shared.ID
	FirstName string
	LastName  string
	Phone     string
}

type Repository interface {
	List(ctx context.Context, filter ListFilter, page, perPage int) (ListResult, error)
	Get(ctx context.Context, id shared.ID) (*Customer, error)
	PhoneExists(ctx context.Context, phone string, excludeID *shared.ID) (bool, error)
	EmailExists(ctx context.Context, email string, excludeID *shared.ID) (bool, error)
	Create(ctx context.Context, customer *Customer) error
	Update(ctx context.Context, customer *Customer) error
	SoftDelete(ctx context.Context, id shared.ID) error
	ListBirthdayToday(ctx context.Context, day time.Time) ([]BirthdayRecipient, error)
	ListPhonesByFilter(ctx context.Context, filter ListFilter) ([]BirthdayRecipient, error)
	WasBirthdaySMSSent(ctx context.Context, customerID shared.ID, day time.Time) (bool, error)
	RecordBirthdaySMS(ctx context.Context, customerID shared.ID, day time.Time, status string, errMsg string) error
}
