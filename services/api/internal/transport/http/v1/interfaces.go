package v1

import (
	"context"
	"hive/services/api/internal/domain/order"
	"hive/services/api/internal/domain/shared"
)

type OrderClient interface {
	CreateOrder(
		ctx context.Context,
		userID string,
		items []string,
		location shared.Location,
	) (*order.OrderInfo, error)
	GetOrder(ctx context.Context, orderID string) (*order.Order, error)
}
