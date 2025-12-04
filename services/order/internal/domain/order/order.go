package order

import (
	"hive/services/order/internal/domain/shared"
	"time"
)

type Order struct {
	ID      string
	UserID  string
	DroneID string

	Items    []string
	Status   OrderStatus
	Location shared.Location

	CreatedAt time.Time
	UpdatedAt time.Time
}
