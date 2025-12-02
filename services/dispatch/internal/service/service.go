package service

import (
	"context"
	"fmt"
	"hive/services/dispatch/internal/domain"
	"time"

	"github.com/google/uuid"
)

type DispatchService struct {
	repo      AssignmentRepository
	order     OrderClient
	tracking  TrackingClient
	telemetry TelemetryClient
}

func NewDispatchService(
	repo AssignmentRepository,
	order OrderClient,
	tracker TrackingClient,
	telemetry TelemetryClient,
) *DispatchService {
	return &DispatchService{
		repo:      repo,
		order:     order,
		tracking:  tracker,
		telemetry: telemetry,
	}
}

func (s *DispatchService) AssignDrone(ctx context.Context, orderID string, deliveryLocation domain.Location) (droneID string, err error) {
	// TODO: StoreService to return nearest store location for delivery
	storeLocation := domain.Location{
		Lat: 55.748281,
		Lon: 37.641499,
	}

	droneID, err = s.tracking.FindNearest(ctx, storeLocation)
	if err != nil {
		return "", fmt.Errorf("failed to find nearest drone: %w", err)
	}

	assignment := &domain.Assignment{
		ID:        uuid.NewString(),
		OrderID:   orderID,
		DroneID:   droneID,
		Status:    domain.AssignmentStatusCreated,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.repo.Save(ctx, assignment); err != nil {
		return "", fmt.Errorf("failed to save assignment: %w", err)
	}

	fail := func(err error) (string, error) {
		assignment.Status = domain.AssignmentStatusFailed
		assignment.UpdatedAt = time.Now().UTC()
		_ = s.repo.Update(ctx, assignment)
		_ = s.tracking.SetStatus(ctx, droneID, domain.DroneStatusFree)
		return "", err
	}

	if err := s.order.UpdateStatus(ctx, orderID, domain.OrderStatusAssigned); err != nil {
		return fail(fmt.Errorf("failed to update order status: %w", err))
	}

	if err := s.tracking.SetStatus(ctx, droneID, domain.DroneStatusBusy); err != nil {
		return fail(fmt.Errorf("failed to set drone status: %w", err))
	}

	if err := s.telemetry.SendCommand(ctx, droneID, domain.DroneActionFlyTo, storeLocation); err != nil {
		return fail(fmt.Errorf("failed to send fly to store command: %w", err))
	}

	assignment.Status = domain.AssignmentStatusAssigned
	assignment.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, assignment); err != nil {
		return fail(fmt.Errorf("failed to update assignment status: %w", err))
	}

	return droneID, nil
}
