package payment

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusPaid       PaymentStatus = "paid"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

type PaymentProvider string

const (
	PaymentProviderZarinpal PaymentProvider = "zarinpal"
	PaymentProviderIDPay    PaymentProvider = "idpay"
	PaymentProviderManual   PaymentProvider = "manual"
	PaymentProviderOther    PaymentProvider = "other"
)

type Payment struct {
	ID         shared.ID
	OrderID    shared.ID
	Provider   PaymentProvider
	ExternalID *string
	Amount     shared.Money
	Status     PaymentStatus
	PaidAt     *time.Time
	Metadata   map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
