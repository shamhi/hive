package service

import (
	"context"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/domain/shared"
)

type DroneRepository interface {
	GetNearest(
		ctx context.Context,
		location shared.Location,
		radiusMeters float64,
		minBattery float64,
	) (*drone.DroneNearest, error)
	GetByID(
		ctx context.Context,
		droneID string,
	) (*drone.Drone, error)
	List(
		ctx context.Context,
		offset, limit int64,
	) ([]*drone.Drone, error)
	SetStatus(
		ctx context.Context,
		droneID string,
		status drone.DroneStatus,
	) error
	UpdateState(
		ctx context.Context,
		tm drone.TelemetryData,
	) error
}
