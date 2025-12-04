package tracking

import (
	"context"
	"fmt"
	pbCommon "hive/gen/common"
	pbTracking "hive/gen/tracking"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/shared"
)

type TrackingClient struct {
	client pbTracking.TrackingServiceClient
}

func NewTrackingClient(client pbTracking.TrackingServiceClient) *TrackingClient {
	return &TrackingClient{client: client}
}

func (c *TrackingClient) FindNearest(
	ctx context.Context,
	storeLocation *shared.Location,
	minBattery float64,
	searchRadius float64,
) (*drone.DroneNearestInfo, error) {
	req := &pbTracking.FindNearestRequest{
		StoreLocation: &pbCommon.Location{
			Lat: storeLocation.Lat,
			Lon: storeLocation.Lon,
		},
		MinBattery:   minBattery,
		RadiusMeters: searchRadius,
	}
	resp, err := c.client.FindNearest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest drone: %w", err)
	}

	if !resp.GetFound() {
		return nil, fmt.Errorf("no available drone found")
	}

	if resp.GetDroneId() == "" {
		return nil, fmt.Errorf("no available drone found")
	}
	if resp.GetDistanceMeters() <= 0 {
		return nil, fmt.Errorf("invalid distance returned for nearest drone")
	}

	return &drone.DroneNearestInfo{
		ID:       resp.GetDroneId(),
		Distance: resp.GetDistanceMeters(),
	}, nil
}

func (c *TrackingClient) GetDroneLocation(ctx context.Context, droneID string) (*drone.DroneInfo, error) {
	req := &pbTracking.GetDroneLocationRequest{
		DroneId: droneID,
	}
	resp, err := c.client.GetDroneLocation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get drone location: %w", err)
	}

	if resp.GetBattery() < 0 || resp.GetBattery() > 100 {
		return nil, fmt.Errorf("invalid battery level returned for drone")
	}
	if resp.GetSpeedMps() < 0 {
		return nil, fmt.Errorf("invalid speed returned for drone")
	}
	if resp.GetConsumptionPerMeter() < 0 {
		return nil, fmt.Errorf("invalid consumption per meter returned for drone")
	}

	locationPb := resp.GetLocation()
	if locationPb == nil {
		return nil, fmt.Errorf("no location data returned for drone")
	}

	return &drone.DroneInfo{
		ID: droneID,
		Location: shared.Location{
			Lat: locationPb.GetLat(),
			Lon: locationPb.GetLon(),
		},
		Battery:             resp.GetBattery(),
		SpeedMps:            resp.GetSpeedMps(),
		ConsumptionPerMeter: resp.GetConsumptionPerMeter(),
	}, nil
}

func (c *TrackingClient) SetStatus(ctx context.Context, droneID string, status drone.DroneStatus) error {
	var newStatus pbTracking.DroneStatus
	switch status {
	case drone.DroneStatusFree:
		newStatus = pbTracking.DroneStatus_STATUS_FREE
	case drone.DroneStatusBusy:
		newStatus = pbTracking.DroneStatus_STATUS_BUSY
	case drone.DroneStatusCharging:
		newStatus = pbTracking.DroneStatus_STATUS_CHARGING
	default:
		newStatus = pbTracking.DroneStatus_STATUS_UNKNOWN
	}

	req := &pbTracking.SetStatusRequest{
		DroneId: droneID,
		Status:  newStatus,
	}
	resp, err := c.client.SetStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set drone status: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("failed to set drone status: operation not successful")
	}

	return nil
}
