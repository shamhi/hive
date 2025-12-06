package service

import (
	"context"
	"hive/services/order/internal/domain/order"
	"hive/services/order/internal/domain/shared"
)

type OrderRepository interface {
	Save(ctx context.Context, o *order.Order) error
	GetByID(ctx context.Context, id string) (*order.Order, error)
	UpdateStatus(ctx context.Context, id string, status order.OrderStatus) error
	SetDroneID(ctx context.Context, id string, droneID string) error
	UpdateDroneAndStatus(ctx context.Context, id string, droneID string, status order.OrderStatus) error
}

type DispatchClient interface {
	AssignDrone(ctx context.Context, orderID string, deliveryLocation *shared.Location) (droneID string, err error)
}
