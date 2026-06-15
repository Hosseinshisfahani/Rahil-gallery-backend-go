package shared

import "github.com/google/uuid"

type ID = uuid.UUID

type Money struct {
	Amount   float64
	Currency string
}
