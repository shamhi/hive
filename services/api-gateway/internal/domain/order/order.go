package order

import (
	"hive/services/api-gateway/internal/domain/shared"
)

type Order struct {
	ID      string
	UserID  string
	DroneID string

	Items    []string
	Status   OrderStatus
	Location shared.Location
}
