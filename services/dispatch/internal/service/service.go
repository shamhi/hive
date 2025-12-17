package service

import (
	"context"
	"errors"
	"fmt"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/order"
	"hive/services/dispatch/internal/domain/shared"
	"net"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
	lg        logger.Logger
}

func NewDispatchService(
	repo AssignmentRepository,
	order OrderClient,
	store StoreClient,
	base BaseClient,
	tracking TrackingClient,
	telemetry TelemetryClient,
	lg logger.Logger,
) *DispatchService {
	return &DispatchService{
		repo:      repo,
		order:     order,
		store:     store,
		base:      base,
		tracking:  tracking,
		telemetry: telemetry,
		lg:        lg,
	}
}

func (s *DispatchService) AssignDrone(
	ctx context.Context,
	orderID string,
	deliveryLocation *shared.Location,
	minDroneBattery float64,
	droneSearchRadius float64,
) (*assignment.AssignmentInfo, error) {
	lg := s.lg.With(
		zap.String("component", "dispatch_service"),
		zap.String("op", "AssignDrone"),
		zap.String("order_id", orderID),
	)

	start := time.Now()
	if orderID == "" {
		lg.Warn(ctx, "validation failed: order ID is empty")
		return nil, fmt.Errorf("order ID is required")
	}
	if deliveryLocation == nil {
		lg.Warn(ctx, "validation failed: delivery location is nil")
		return nil, fmt.Errorf("delivery location is required")
	}
	if minDroneBattery < 0 {
		lg.Warn(ctx, "validation failed: minDroneBattery < 0", zap.Float64("min_drone_battery", minDroneBattery))
		return nil, fmt.Errorf("min drone battery must be non-negative")
	}
	if droneSearchRadius <= 0 {
		lg.Warn(ctx, "validation failed: droneSearchRadius <= 0", zap.Float64("drone_search_radius", droneSearchRadius))
		return nil, fmt.Errorf("drone search radius must be positive")
	}

	lg.Info(ctx, "assign started",
		zap.Float64("delivery_lat", deliveryLocation.Lat),
		zap.Float64("delivery_lon", deliveryLocation.Lon),
		zap.Float64("min_drone_battery", minDroneBattery),
		zap.Float64("drone_search_radius", droneSearchRadius),
	)

	nearestStore, err := s.store.FindNearest(ctx, deliveryLocation)
	if err != nil {
		lg.Error(ctx, "failed to find nearest store", zap.Error(err))
		return nil, fmt.Errorf("failed to find nearest store for order %s: %w", orderID, err)
	}
	lg.Info(ctx, "nearest store selected", zap.String("store_id", nearestStore.ID), zap.Float64("store_distance", nearestStore.Distance))

	storeInfo, err := s.store.GetStoreLocation(ctx, nearestStore.ID)
	if err != nil {
		lg.Error(ctx, "failed to get store location", zap.String("store_id", nearestStore.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to get store location %s for order %s: %w", nearestStore.ID, orderID, err)
	}
	lg.Info(ctx, "store location loaded",
		zap.String("store_id", nearestStore.ID),
		zap.Float64("store_lat", storeInfo.Location.Lat),
		zap.Float64("store_lon", storeInfo.Location.Lon),
	)

	nearestDrone, err := s.tracking.FindNearest(ctx, &storeInfo.Location, minDroneBattery, droneSearchRadius)
	if err != nil {
		lg.Error(ctx, "failed to find nearest drone", zap.Error(err))
		return nil, fmt.Errorf("failed to find nearest drone for order %s: %w", orderID, err)
	}

	droneID := nearestDrone.ID
	lg = lg.With(zap.String("drone_id", droneID))
	lg.Info(ctx, "nearest drone selected", zap.Float64("drone_distance_to_store", nearestDrone.Distance))

	droneInfo, err := s.tracking.GetDroneLocation(ctx, droneID)
	if err != nil {
		lg.Error(ctx, "failed to get drone location", zap.Error(err))
		return nil, fmt.Errorf("failed to get drone location for drone %s: %w", droneID, err)
	}
	lg.Info(ctx, "drone info loaded",
		zap.Float64("drone_battery", droneInfo.Battery),
		zap.Float64("drone_speed_mps", droneInfo.SpeedMps),
		zap.Float64("drone_consumption_per_meter", droneInfo.ConsumptionPerMeter),
		zap.String("drone_status", string(droneInfo.Status)),
	)

	nearestBase, err := s.base.FindNearest(ctx, deliveryLocation)
	if err != nil {
		lg.Error(ctx, "failed to find nearest base", zap.Error(err))
		return nil, fmt.Errorf("failed to find nearest base: %w", err)
	}
	lg.Info(ctx, "nearest base selected", zap.String("base_id", nearestBase.ID), zap.Float64("base_distance", nearestBase.Distance))

	totalDistance := nearestDrone.Distance + nearestStore.Distance + nearestBase.Distance
	batteryRequired := totalDistance * droneInfo.ConsumptionPerMeter

	lg.Info(ctx, "battery check",
		zap.Float64("total_distance", totalDistance),
		zap.Float64("battery_required", batteryRequired),
		zap.Float64("battery_available", droneInfo.Battery),
	)

	if batteryRequired > droneInfo.Battery {
		lg.Warn(ctx, "not enough battery",
			zap.Float64("battery_required", batteryRequired),
			zap.Float64("battery_available", droneInfo.Battery),
		)
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
		lg.Error(ctx, "failed to save assignment", zap.String("assignment_id", a.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to save assignment: %w", err)
	}
	lg.Info(ctx, "assignment created", zap.String("assignment_id", a.ID), zap.String("status", string(a.Status)))

	fail := func(cause error) (*assignment.AssignmentInfo, error) {
		lg.Error(ctx, "assign failed, starting rollback",
			zap.String("assignment_id", a.ID),
			zap.Error(cause),
		)

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFailed); err != nil {
			lg.Warn(ctx, "rollback: failed to set assignment status FAILED", zap.String("assignment_id", a.ID), zap.Error(err))
		} else {
			lg.Info(ctx, "rollback: assignment marked as FAILED", zap.String("assignment_id", a.ID))
		}

		if err := s.tracking.SetStatus(ctx, droneID, drone.DroneStatusFree); err != nil {
			lg.Warn(ctx, "rollback: failed to set drone status FREE", zap.Error(err))
		} else {
			lg.Info(ctx, "rollback: drone marked as FREE")
		}

		return nil, cause
	}

	if err := s.tracking.SetStatus(ctx, droneID, drone.DroneStatusBusy); err != nil {
		lg.Error(ctx, "failed to set drone status BUSY", zap.Error(err))
		return fail(fmt.Errorf("failed to set drone status: %w", err))
	}
	lg.Info(ctx, "drone marked as BUSY")

	if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAssigned); err != nil {
		lg.Error(ctx, "failed to update assignment status to ASSIGNED", zap.String("assignment_id", a.ID), zap.Error(err))
		return fail(fmt.Errorf("failed to update assignment status: %w", err))
	}
	lg.Info(ctx, "assignment marked as ASSIGNED", zap.String("assignment_id", a.ID))

	target := &drone.Target{Location: &storeInfo.Location, Type: drone.TargetTypeStore}
	lg.Info(ctx, "sending command: FLY_TO store",
		zap.String("target_type", string(target.Type)),
		zap.Float64("target_lat", target.Location.Lat),
		zap.Float64("target_lon", target.Location.Lon),
	)

	if err := s.telemetry.SendCommand(ctx, droneID, drone.DroneActionFlyTo, target); err != nil {
		lg.Error(ctx, "failed to send FLY_TO store command", zap.Error(err))
		return fail(fmt.Errorf("failed to send fly to store command: %w", err))
	}

	if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFlyingToStore); err != nil {
		lg.Error(ctx, "failed to update assignment status to FLYING_TO_STORE", zap.String("assignment_id", a.ID), zap.Error(err))
		return fail(fmt.Errorf("failed to update assignment status: %w", err))
	}
	lg.Info(ctx, "assignment marked as FLYING_TO_STORE", zap.String("assignment_id", a.ID))

	if droneInfo.SpeedMps <= 0 {
		lg.Error(ctx, "invalid drone speed", zap.Float64("speed_mps", droneInfo.SpeedMps))
		return fail(fmt.Errorf("invalid drone %s speed: %.2fm/s", droneID, droneInfo.SpeedMps))
	}

	eta := int32(totalDistance / droneInfo.SpeedMps)
	lg.Info(ctx, "assign completed",
		zap.Int32("eta_seconds", eta),
		zap.Duration("duration", time.Since(start)),
	)

	return &assignment.AssignmentInfo{
		DroneID:    droneID,
		EtaSeconds: eta,
	}, nil
}

func (s *DispatchService) GetAssignment(
	ctx context.Context,
	droneID string,
) (*assignment.Assignment, error) {
	lg := s.lg.With(
		zap.String("component", "dispatch_service"),
		zap.String("op", "GetAssignment"),
		zap.String("drone_id", droneID),
	)

	start := time.Now()
	a, err := s.repo.GetByDroneID(ctx, droneID)
	if err != nil {
		lg.Error(ctx, "failed to get assignment by drone ID", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get assignment by drone ID %s: %w", droneID, err)
	}

	lg.Info(ctx, "get assignment completed",
		zap.String("assignment_id", a.ID),
		zap.String("order_id", a.OrderID),
		zap.String("status", string(a.Status)),
		zap.Duration("duration", time.Since(start)),
	)

	return a, nil
}

func (s *DispatchService) HandleTelemetryEvent(ctx context.Context, event drone.TelemetryEvent) error {
	lg := s.lg.With(
		zap.String("component", "dispatch_service"),
		zap.String("op", "HandleTelemetryEvent"),
		zap.String("drone_id", event.DroneID),
		zap.String("event", string(event.Event)),
	)

	start := time.Now()
	lg.Info(ctx, "telemetry event received",
		zap.Float64("drone_lat", event.DroneLocation.Lat),
		zap.Float64("drone_lon", event.DroneLocation.Lon),
		zap.Int64("timestamp", event.Timestamp),
	)

	if event.Event == drone.DroneEventFullyCharged {
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.tracking.SetStatus(ctx, event.DroneID, drone.DroneStatusFree)
		}); err != nil {
			lg.Error(ctx, "failed to set drone status FREE after retries", zap.Error(err))
			return fmt.Errorf("failed to set drone status to FREE after retries: %w", err)
		}
		lg.Info(ctx, "drone marked as FREE (fully charged)", zap.Duration("duration", time.Since(start)))
		return nil
	}

	a, err := s.repo.GetByDroneID(ctx, event.DroneID)
	if err != nil {
		if errors.Is(err, ErrAssignmentNotFound) {
			lg.Info(ctx, "no assignment found for drone, ignoring event", zap.Duration("duration", time.Since(start)))
			return nil
		}
		lg.Error(ctx, "failed to get assignment by drone ID", zap.Error(err))
		return fmt.Errorf("failed to get assignment by drone ID %s: %w", event.DroneID, err)
	}

	lg = lg.With(
		zap.String("assignment_id", a.ID),
		zap.String("order_id", a.OrderID),
		zap.String("assignment_status", string(a.Status)),
	)

	lg.Info(ctx, "assignment loaded for telemetry event")

	switch event.Event {
	case drone.DroneEventArrivedAtStore:
		if a.Status == assignment.AssignmentStatusAtStore ||
			a.Status == assignment.AssignmentStatusPickedUpCargo ||
			a.Status == assignment.AssignmentStatusFlyingToClient ||
			a.Status == assignment.AssignmentStatusAtClient ||
			a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			lg.Info(ctx, "event ignored due to current assignment status", zap.Duration("duration", time.Since(start)))
			return nil
		}

		lg.Info(ctx, "transition: AT_STORE")
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAtStore); err != nil {
			lg.Error(ctx, "failed to update assignment status to AT_STORE", zap.Error(err))
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		lg.Info(ctx, "sending command: PICKUP_CARGO")
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionPickupCargo, nil)
		}); err != nil {
			lg.Error(ctx, "failed to send pickup cargo command after retries", zap.Error(err))
			return fmt.Errorf("failed to send pickup cargo command after retries: %w", err)
		}

		lg.Info(ctx, "event processed", zap.Duration("duration", time.Since(start)))
		return nil

	case drone.DroneEventPickedUpCargo:
		if a.Status == assignment.AssignmentStatusPickedUpCargo ||
			a.Status == assignment.AssignmentStatusFlyingToClient ||
			a.Status == assignment.AssignmentStatusAtClient ||
			a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			lg.Info(ctx, "event ignored due to current assignment status", zap.Duration("duration", time.Since(start)))
			return nil
		}

		lg.Info(ctx, "transition: PICKED_UP_CARGO")
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusPickedUpCargo); err != nil {
			lg.Error(ctx, "failed to update assignment status to PICKED_UP_CARGO", zap.Error(err))
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if a.Target == nil {
			lg.Error(ctx, "assignment target location is nil")
			return fmt.Errorf("assignment target location is nil")
		}

		target := &drone.Target{Location: a.Target, Type: drone.TargetTypeClient}
		lg.Info(ctx, "sending command: FLY_TO client",
			zap.String("target_type", string(target.Type)),
			zap.Float64("target_lat", target.Location.Lat),
			zap.Float64("target_lon", target.Location.Lon),
		)

		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionFlyTo, target)
		}); err != nil {
			lg.Error(ctx, "failed to send fly to client command after retries", zap.Error(err))
			return fmt.Errorf("failed to send fly to client command after retries: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFlyingToClient); err != nil {
			lg.Error(ctx, "failed to update assignment status to FLYING_TO_CLIENT", zap.Error(err))
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		lg.Info(ctx, "event processed", zap.Duration("duration", time.Since(start)))
		return nil

	case drone.DroneEventArrivedAtClient:
		if a.Status == assignment.AssignmentStatusAtClient ||
			a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			lg.Info(ctx, "event ignored due to current assignment status", zap.Duration("duration", time.Since(start)))
			return nil
		}

		lg.Info(ctx, "transition: AT_CLIENT")
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAtClient); err != nil {
			lg.Error(ctx, "failed to update assignment status to AT_CLIENT", zap.Error(err))
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		lg.Info(ctx, "sending command: DROP_CARGO")
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionDropCargo, nil)
		}); err != nil {
			lg.Error(ctx, "failed to send drop cargo command after retries", zap.Error(err))
			return fmt.Errorf("failed to send drop cargo command after retries: %w", err)
		}

		lg.Info(ctx, "event processed", zap.Duration("duration", time.Since(start)))
		return nil

	case drone.DroneEventDroppedCargo:
		if a.Status == assignment.AssignmentStatusDroppedCargo ||
			a.Status == assignment.AssignmentStatusReturningBase ||
			a.Status == assignment.AssignmentStatusCompleted {
			lg.Info(ctx, "event ignored due to current assignment status", zap.Duration("duration", time.Since(start)))
			return nil
		}

		lg.Info(ctx, "transition: DROPPED_CARGO")
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusDroppedCargo); err != nil {
			lg.Error(ctx, "failed to update assignment status to DROPPED_CARGO", zap.Error(err))
			return fmt.Errorf("failed to update assignment %s status to DROPPED_CARGO: %w", a.ID, err)
		}

		lg.Info(ctx, "updating order status to COMPLETED (best-effort)")
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.order.UpdateStatus(ctx, a.OrderID, order.OrderStatusCompleted)
		}); err != nil {
			lg.Warn(ctx, "failed to update order status to COMPLETED, continuing", zap.Error(err))
		}

		var baseID string
		lg.Info(ctx, "finding nearest base")
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			nb, err := s.base.FindNearest(ctx, &event.DroneLocation)
			if err != nil {
				return err
			}
			baseID = nb.ID
			return nil
		}); err != nil {
			lg.Error(ctx, "failed to find nearest base after retries", zap.Error(err))
			return fmt.Errorf("failed to find nearest base after retries: %w", err)
		}
		lg.Info(ctx, "nearest base selected", zap.String("base_id", baseID))

		var baseLoc shared.Location
		lg.Info(ctx, "loading base location", zap.String("base_id", baseID))
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			b, err := s.base.GetBaseLocation(ctx, baseID)
			if err != nil {
				return err
			}
			baseLoc = b.Location
			return nil
		}); err != nil {
			lg.Error(ctx, "failed to get base location after retries", zap.Error(err))
			return fmt.Errorf("failed to get base location after retries: %w", err)
		}

		target := &drone.Target{Location: &baseLoc, Type: drone.TargetTypeBase}
		lg.Info(ctx, "sending command: FLY_TO base",
			zap.String("target_type", string(target.Type)),
			zap.Float64("target_lat", target.Location.Lat),
			zap.Float64("target_lon", target.Location.Lon),
		)

		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionFlyTo, target)
		}); err != nil {
			lg.Error(ctx, "failed to send fly to base command after retries", zap.Error(err))
			return fmt.Errorf("failed to send fly to base command after retries: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusReturningBase); err != nil {
			lg.Error(ctx, "failed to update assignment status to RETURNING_BASE", zap.Error(err))
			return fmt.Errorf("failed to update assignment status to RETURNING_BASE: %w", err)
		}

		lg.Info(ctx, "event processed", zap.Duration("duration", time.Since(start)))
		return nil

	case drone.DroneEventArrivedAtBase:
		if a.Status == assignment.AssignmentStatusCompleted {
			lg.Info(ctx, "event ignored due to current assignment status", zap.Duration("duration", time.Since(start)))
			return nil
		}

		lg.Info(ctx, "sending command: CHARGE")
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionCharge, nil)
		}); err != nil {
			lg.Error(ctx, "failed to send charge command after retries", zap.Error(err))
			return fmt.Errorf("failed to send charge command after retries: %w", err)
		}

		lg.Info(ctx, "setting drone status: CHARGING")
		if err := resilience.Retry(ctx, retryCfg, shouldRetryDefault, func(ctx context.Context) error {
			return s.tracking.SetStatus(ctx, event.DroneID, drone.DroneStatusCharging)
		}); err != nil {
			lg.Error(ctx, "failed to set drone status to CHARGING after retries", zap.Error(err))
			return fmt.Errorf("failed to set drone status to CHARGING after retries: %w", err)
		}

		lg.Info(ctx, "transition: COMPLETED")
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusCompleted); err != nil {
			lg.Error(ctx, "failed to update assignment status to COMPLETED", zap.Error(err))
			return fmt.Errorf("failed to update assignment status to COMPLETED: %w", err)
		}

		lg.Info(ctx, "event processed", zap.Duration("duration", time.Since(start)))
		return nil

	default:
		lg.Warn(ctx, "unknown telemetry event", zap.Duration("duration", time.Since(start)))
		return fmt.Errorf("unknown telemetry event: %s", event.Event)
	}
}
