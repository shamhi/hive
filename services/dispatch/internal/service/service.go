package service

import (
	"context"
	"errors"
	"fmt"
	"hive/pkg/resilience"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/order"
	"hive/services/dispatch/internal/domain/shared"
	"net"
	"time"

	"github.com/google/uuid"
)

var retryCfg = resilience.RetryConfig{
	MaxAttempts: 4,
	BaseDelay:   80 * time.Millisecond,
	MaxDelay:    800 * time.Millisecond,
	Jitter:      0.2,
}

func shouldRetryDefault(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var ne net.Error
	if errors.As(err, &ne) {
		return true
	}

	return true
}

type DispatchService struct {
	repo      AssignmentRepository
	order     OrderClient
	store     StoreClient
	base      BaseClient
	tracking  TrackingClient
	telemetry TelemetryClient
}

func NewDispatchService(
	repo AssignmentRepository,
	order OrderClient,
	store StoreClient,
	base BaseClient,
	tracking TrackingClient,
	telemetry TelemetryClient,
) *DispatchService {
	return &DispatchService{
		repo:      repo,
		order:     order,
		store:     store,
		base:      base,
		tracking:  tracking,
		telemetry: telemetry,
	}
}

func (s *DispatchService) AssignDrone(
	ctx context.Context,
	orderID string,
	deliveryLocation *shared.Location,
	minDroneBattery float64,
	droneSearchRadius float64,
) (*assignment.AssignmentInfo, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if deliveryLocation == nil {
		return nil, fmt.Errorf("delivery location is required")
	}
	if minDroneBattery < 0 {
		return nil, fmt.Errorf("min drone battery must be non-negative")
	}
	if droneSearchRadius <= 0 {
		return nil, fmt.Errorf("drone search radius must be positive")
	}

	nearestStore, err := s.store.FindNearest(ctx, deliveryLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest store for order %s: %w", orderID, err)
	}

	storeInfo, err := s.store.GetStoreLocation(ctx, nearestStore.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get store location %s for order %s: %w", nearestStore.ID, orderID, err)
	}

	nearestDrone, err := s.tracking.FindNearest(ctx, &storeInfo.Location, minDroneBattery, droneSearchRadius)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest drone for order %s: %w", orderID, err)
	}

	droneID := nearestDrone.ID

	droneInfo, err := s.tracking.GetDroneLocation(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get drone location for drone %s: %w", droneID, err)
	}

	nearestBase, err := s.base.FindNearest(ctx, deliveryLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest base: %w", err)
	}

	totalDistance := nearestDrone.Distance + nearestStore.Distance + nearestBase.Distance
	batteryRequired := totalDistance * droneInfo.ConsumptionPerMeter
	if batteryRequired > droneInfo.Battery {
		return nil, fmt.Errorf("drone %s does not have enough battery for the trip", droneID)
	}

	a := &assignment.Assignment{
		ID:      uuid.NewString(),
		OrderID: orderID,
		DroneID: droneID,
		Status:  assignment.AssignmentStatusCreated,
		Target: &shared.Location{
			Lat: deliveryLocation.Lat,
			Lon: deliveryLocation.Lon,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.repo.Save(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to save assignment: %w", err)
	}

	fail := func(cause error) (*assignment.AssignmentInfo, error) {
		_ = s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFailed)
		_ = s.tracking.SetStatus(ctx, droneID, drone.DroneStatusFree)
		return nil, cause
	}

	if err := s.tracking.SetStatus(ctx, droneID, drone.DroneStatusBusy); err != nil {
		return fail(fmt.Errorf("failed to set drone status: %w", err))
	}

	if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAssigned); err != nil {
		return fail(fmt.Errorf("failed to update assignment status: %w", err))
	}

	target := &drone.Target{Location: &storeInfo.Location, Type: drone.TargetTypeStore}
	if err := s.telemetry.SendCommand(ctx, droneID, drone.DroneActionFlyTo, target); err != nil {
		return fail(fmt.Errorf("failed to send fly to store command: %w", err))
	}

	if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFlyingToStore); err != nil {
		return fail(fmt.Errorf("failed to update assignment status: %w", err))
	}

	if droneInfo.SpeedMps <= 0 {
		return fail(fmt.Errorf("invalid drone %s speed: %.2fm/s", droneID, droneInfo.SpeedMps))
	}
	eta := int32(totalDistance / droneInfo.SpeedMps)

	return &assignment.AssignmentInfo{
		DroneID:    droneID,
		EtaSeconds: eta,
	}, nil
}

func (s *DispatchService) GetAssignment(
	ctx context.Context,
	droneID string,
) (*assignment.Assignment, error) {
	a, err := s.repo.GetByDroneID(ctx, droneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment by drone ID %s: %w", droneID, err)
	}

	return a, nil
}

func (s *DispatchService) HandleTelemetryEvent(ctx context.Context, event drone.TelemetryEvent) error {
	if event.Event == drone.DroneEventFullyCharged {
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.tracking.SetStatus(ctx, event.DroneID, drone.DroneStatusFree)
		}); err != nil {
			return fmt.Errorf("failed to set drone status to FREE after retries: %w", err)
		}
		return nil
	}

	a, err := s.repo.GetByDroneID(ctx, event.DroneID)
	if err != nil {
		if errors.Is(err, ErrAssignmentNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get assignment by drone ID %s: %w", event.DroneID, err)
	}

	switch event.Event {
	case drone.DroneEventArrivedAtStore:
		if a.Status == assignment.AssignmentStatusAtStore ||
			a.Status == assignment.AssignmentStatusPickedUpCargo ||
			a.Status == assignment.AssignmentStatusFlyingToClient ||
			a.Status == assignment.AssignmentStatusAtClient ||
			a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			return nil
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAtStore); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionPickupCargo, nil)
		}); err != nil {
			return fmt.Errorf("failed to send pickup cargo command after retries: %w", err)
		}

		return nil

	case drone.DroneEventPickedUpCargo:
		if a.Status == assignment.AssignmentStatusPickedUpCargo ||
			a.Status == assignment.AssignmentStatusFlyingToClient ||
			a.Status == assignment.AssignmentStatusAtClient ||
			a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			return nil
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusPickedUpCargo); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if a.Target == nil {
			return fmt.Errorf("assignment target location is nil")
		}

		target := &drone.Target{Location: a.Target, Type: drone.TargetTypeClient}
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionFlyTo, target)
		}); err != nil {
			return fmt.Errorf("failed to send fly to client command after retries: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFlyingToClient); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		return nil

	case drone.DroneEventArrivedAtClient:
		if a.Status == assignment.AssignmentStatusAtClient ||
			a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			return nil
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAtClient); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionDropCargo, nil)
		}); err != nil {
			return fmt.Errorf("failed to send drop cargo command after retries: %w", err)
		}

		return nil

	case drone.DroneEventDroppedCargo:
		if a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			return nil
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusDroppedCargo); err != nil {
			return fmt.Errorf("failed to update assignment %s status to DROPPED_CARGO: %w", a.ID, err)
		}

		_ = resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.order.UpdateStatus(ctx, a.OrderID, order.OrderStatusCompleted)
		})

		var baseID string
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			nb, err := s.base.FindNearest(ctx, &event.DroneLocation)
			if err != nil {
				return err
			}
			baseID = nb.ID
			return nil
		}); err != nil {
			return fmt.Errorf("failed to find nearest base after retries: %w", err)
		}

		var baseLoc shared.Location
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			b, err := s.base.GetBaseLocation(ctx, baseID)
			if err != nil {
				return err
			}
			baseLoc = b.Location
			return nil
		}); err != nil {
			return fmt.Errorf("failed to get base location after retries: %w", err)
		}

		target := &drone.Target{Location: &baseLoc, Type: drone.TargetTypeBase}
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionFlyTo, target)
		}); err != nil {
			return fmt.Errorf("failed to send fly to base command after retries: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusReturningBase); err != nil {
			return fmt.Errorf("failed to update assignment status to RETURNING_BASE: %w", err)
		}

		return nil

	case drone.DroneEventArrivedAtBase:
		if a.Status == assignment.AssignmentStatusCompleted {
			return nil
		}

		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionCharge, nil)
		}); err != nil {
			return fmt.Errorf("failed to send charge command after retries: %w", err)
		}

		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.tracking.SetStatus(ctx, event.DroneID, drone.DroneStatusCharging)
		}); err != nil {
			return fmt.Errorf("failed to set drone status to CHARGING after retries: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusCompleted); err != nil {
			return fmt.Errorf("failed to update assignment status to COMPLETED: %w", err)
		}

		return nil

	default:
		return fmt.Errorf("unknown telemetry event: %s", event.Event)
	}
}
