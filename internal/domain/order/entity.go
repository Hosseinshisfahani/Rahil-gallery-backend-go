package order

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type Order struct {
	ID                 shared.ID
	OrderNumber        string
	UserID             shared.ID
	Status             OrderStatus
	Subtotal           shared.Money
	DiscountAmount     shared.Money
	ShippingAmount     shared.Money
	TaxAmount          shared.Money
	TotalAmount        shared.Money
	CustomerNote       *string
	ShippingAddressID  *shared.ID
	BillingAddressID   *shared.ID
	Items              []OrderItem
	PlacedAt           time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OrderItem struct {
	ID           shared.ID
	OrderID      shared.ID
	VariantID    shared.ID
	ProductName  string
	VariantName  string
	SKU          string
	JewelryType  catalog.JewelryType
	MetalType    *catalog.MetalType
	Karat        *int16
	Quantity     int
	UnitPrice    shared.Money
	LineTotal    shared.Money
	WeightGrams  *float64
}

type OrderStatusHistory struct {
	ID         shared.ID
	OrderID    shared.ID
	FromStatus *OrderStatus
	ToStatus   OrderStatus
	Note       *string
	ChangedBy  *shared.ID
	CreatedAt  time.Time
}
