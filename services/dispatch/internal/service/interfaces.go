package service

import (
	"context"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/base"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/order"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/domain/store"
)

type AssignmentRepository interface {
	Save(ctx context.Context, a *assignment.Assignment) error
	GetByID(ctx context.Context, id string) (*assignment.Assignment, error)
	GetByDroneID(ctx context.Context, droneID string) (*assignment.Assignment, error)
	UpdateStatus(ctx context.Context, id string, status assignment.AssignmentStatus) error
}

type TrackingClient interface {
	FindNearest(ctx context.Context, storeLocation *shared.Location, minBattery, radius float64) (*drone.DroneNearest, error)
	GetDroneLocation(ctx context.Context, droneID string) (*drone.Drone, error)
	SetStatus(ctx context.Context, droneID string, status drone.DroneStatus) error
}

type TelemetryClient interface {
	SendCommand(ctx context.Context, droneID string, action drone.DroneAction, target *drone.Target) error
}

type StoreClient interface {
	FindNearest(ctx context.Context, deliveryLocation *shared.Location) (*store.StoreNearest, error)
	GetStoreLocation(ctx context.Context, storeID string) (*store.Store, error)
}

type BaseClient interface {
	FindNearest(ctx context.Context, deliveryLocation *shared.Location) (*base.BaseNearest, error)
	GetBaseLocation(ctx context.Context, baseID string) (*base.Base, error)
}

type OrderClient interface {
	UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error
}
