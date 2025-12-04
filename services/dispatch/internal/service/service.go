package service

import (
	"context"
	"fmt"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/order"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/utils"
	"time"

	"github.com/google/uuid"
)

type DispatchService struct {
	repo      AssignmentRepository
	order     OrderClient
	store     StoreClient
	tracking  TrackingClient
	telemetry TelemetryClient
}

func NewDispatchService(
	repo AssignmentRepository,
	order OrderClient,
	store StoreClient,
	tracking TrackingClient,
	telemetry TelemetryClient,
) *DispatchService {
	return &DispatchService{
		repo:      repo,
		order:     order,
		store:     store,
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
) (droneID string, err error) {
	if orderID == "" {
		return "", fmt.Errorf("order ID is required")
	}
	if deliveryLocation == nil {
		return "", fmt.Errorf("delivery location is required")
	}

	nearestStore, err := s.store.FindNearest(ctx, deliveryLocation)
	if err != nil {
		return "", fmt.Errorf("failed to find nearest store: %w", err)
	}

	storeInfo, err := s.store.GetStoreLocation(ctx, nearestStore.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get store location: %w", err)
	}

	nearestDrone, err := s.tracking.FindNearest(ctx, &storeInfo.Location, minDroneBattery, droneSearchRadius)
	if err != nil {
		return "", fmt.Errorf("failed to find nearest drone: %w", err)
	}

	droneID = nearestDrone.ID

	droneInfo, err := s.tracking.GetDroneLocation(ctx, droneID)
	if err != nil {
		return "", fmt.Errorf("failed to get drone location: %w", err)
	}

	baseLocation := &shared.Location{Lat: 54.9813120, Lon: 37.1938342}                                          // TODO: Base service for finding nearest available base
	totalDistance := nearestDrone.Distance + nearestStore.Distance + utils.Dist(deliveryLocation, baseLocation) // TODO: nearestBase.Distance
	batteryRequired := totalDistance * droneInfo.ConsumptionPerMeter
	if batteryRequired > droneInfo.Battery {
		return "", fmt.Errorf("drone %s does not have enough battery for the trip", droneID)
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
		return "", fmt.Errorf("failed to save assignment: %w", err)
	}

	fail := func(err error) (string, error) {
		_ = s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFailed)
		_ = s.tracking.SetStatus(ctx, droneID, drone.DroneStatusFree)
		return "", err
	}

	if err := s.order.UpdateStatus(ctx, orderID, order.OrderStatusAssigned); err != nil {
		return fail(fmt.Errorf("failed to update order status: %w", err))
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

	return droneID, nil
}

func (s *DispatchService) HandleTelemetryEvent(ctx context.Context, event drone.TelemetryEvent) error {
	a, err := s.repo.GetByDroneID(ctx, event.DroneID)
	if err != nil {
		return fmt.Errorf("failed to get assignment by drone ID: %w", err)
	}

	switch event.Event {
	case drone.DroneEventArrivedAtStore:
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAtStore); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if err := s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionPickupCargo, nil); err != nil {
			return fmt.Errorf("failed to send pickup cargo command: %w", err)
		}

		return nil
	case drone.DroneEventPickedUpCargo:
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusPickedUpCargo); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		target := &drone.Target{Location: a.Target, Type: drone.TargetTypeClient}
		if err := s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionFlyTo, target); err != nil {
			return fmt.Errorf("failed to send fly to client command: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusFlyingToClient); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		return nil
	case drone.DroneEventArrivedAtClient:
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusAtClient); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if err := s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionDropCargo, nil); err != nil {
			return fmt.Errorf("failed to send drop cargo command: %w", err)
		}

		return nil
	case drone.DroneEventDroppedCargo:
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusDroppedCargo); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if err := s.order.UpdateStatus(ctx, a.OrderID, order.OrderStatusCompleted); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		baseLocation := &shared.Location{Lat: 0, Lon: 0} // TODO: Base service to get actual base location
		target := &drone.Target{Location: baseLocation, Type: drone.TargetTypeBase}
		if err := s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionFlyTo, target); err != nil {
			return fmt.Errorf("failed to send fly to base command: %w", err)
		}

		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusReturningBase); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		return nil
	case drone.DroneEventArrivedAtBase:
		if err := s.repo.UpdateStatus(ctx, a.ID, assignment.AssignmentStatusCompleted); err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		if err := s.telemetry.SendCommand(ctx, event.DroneID, drone.DroneActionCharge, nil); err != nil {
			return fmt.Errorf("failed to send charge command: %w", err)
		}

		if err := s.tracking.SetStatus(ctx, event.DroneID, drone.DroneStatusCharging); err != nil {
			return fmt.Errorf("failed to set drone status to CHARGING: %w", err)
		}

		return nil
	case drone.DroneEventFullyCharged:
		if err := s.tracking.SetStatus(ctx, event.DroneID, drone.DroneStatusFree); err != nil {
			return fmt.Errorf("failed to set drone status to FREE: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unknown telemetry event: %s", event.Event)
	}
}
