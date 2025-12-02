package service

import (
	"context"
	"hive/services/dispatch/internal/domain"
)

type AssignmentRepository interface {
	Save(ctx context.Context, assignment *domain.Assignment) error
	Get(ctx context.Context, id string) (*domain.Assignment, error)
	Update(ctx context.Context, assignment *domain.Assignment) error
}

type TrackingClient interface {
	FindNearest(ctx context.Context, storeLocation domain.Location) (droneID string, err error)
	SetStatus(ctx context.Context, droneID string, status domain.DroneStatus) error
}

type TelemetryClient interface {
	SendCommand(ctx context.Context, droneID string, action domain.DroneAction, target domain.Location) error
}

type OrderClient interface {
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
}
