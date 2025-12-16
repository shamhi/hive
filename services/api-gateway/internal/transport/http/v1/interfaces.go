package v1

import (
	"context"
	"hive/services/api-gateway/internal/domain/base"
	"hive/services/api-gateway/internal/domain/drone"
	"hive/services/api-gateway/internal/domain/order"
	"hive/services/api-gateway/internal/domain/shared"
	"hive/services/api-gateway/internal/domain/store"
)

type OrderClient interface {
	CreateOrder(
		ctx context.Context,
		userID string,
		items []string,
		location shared.Location,
	) (*order.OrderInfo, error)
	GetOrder(
		ctx context.Context,
		orderID string,
	) (*order.Order, error)
}

type BaseClient interface {
	ListBases(
		ctx context.Context,
		offset, limit int64,
	) ([]*base.Base, error)
}

type StoreClient interface {
	ListStores(
		ctx context.Context,
		offset, limit int64,
	) ([]*store.Store, error)
}

type TrackingClient interface {
	ListDrones(
		ctx context.Context,
		offset, limit int64,
	) ([]*drone.Drone, error)
}
