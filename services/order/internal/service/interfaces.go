package service

import (
	"context"
	"hive/services/order/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	Get(ctx context.Context, ID string) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
}

type DispatchClient interface {
	AssignDrone(ctx context.Context, orderID string, loc domain.Location) (droneID string, err error)
}
