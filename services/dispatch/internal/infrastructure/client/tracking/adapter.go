package tracking

import (
	"context"
	"fmt"
	pbTracking "hive/gen/tracking"
	"hive/services/dispatch/internal/domain"
)

type TrackingAdapter struct {
	client pbTracking.TrackingServiceClient
}

func NewTrackingAdapter(client pbTracking.TrackingServiceClient) *TrackingAdapter {
	return &TrackingAdapter{client: client}
}

func (a *TrackingAdapter) FindNearest(ctx context.Context, storeLocation domain.Location) (string, error) {
	req := &pbTracking.FindNearestRequest{
		StoreLocation: &pbTracking.Location{
			Lat: storeLocation.Lat,
			Lon: storeLocation.Lon,
		},
	}
	resp, err := a.client.FindNearest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to find nearest drone: %w", err)
	}

	if !resp.GetFound() {
		return "", fmt.Errorf("no available drone found")
	}

	droneID := resp.GetDroneId()
	if droneID == "" {
		return "", fmt.Errorf("no drone ID returned")
	}

	return droneID, nil
}

func (a *TrackingAdapter) SetStatus(ctx context.Context, droneID string, status domain.DroneStatus) error {
	var newStatus pbTracking.DroneStatus
	switch status {
	case domain.DroneStatusFree:
		newStatus = pbTracking.DroneStatus_STATUS_FREE
	case domain.DroneStatusBusy:
		newStatus = pbTracking.DroneStatus_STATUS_BUSY
	case domain.DroneStatusCharging:
		newStatus = pbTracking.DroneStatus_STATUS_CHARGING
	default:
		newStatus = pbTracking.DroneStatus_STATUS_UNKNOWN
	}

	req := &pbTracking.SetStatusRequest{
		DroneId: droneID,
		Status:  newStatus,
	}
	resp, err := a.client.SetStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set drone status: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("failed to set drone status: operation not successful")
	}

	return nil
}
