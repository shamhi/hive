package dispatch

import (
	"context"
	"fmt"
	pb "hive/gen/dispatch"
	"hive/services/order/internal/domain"
)

type DispatchAdapter struct {
	client pb.DispatchServiceClient
}

func NewDispatchAdapter(client pb.DispatchServiceClient) *DispatchAdapter {
	return &DispatchAdapter{client: client}
}

func (a *DispatchAdapter) AssignDrone(ctx context.Context, id string, loc domain.Location) (string, error) {
	req := &pb.AssignDroneRequest{
		OrderId: id,
		DeliveryLocation: &pb.Location{
			Lat: loc.Lat,
			Lon: loc.Lon,
		},
	}
	resp, err := a.client.AssignDrone(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to assign drone: %w", err)
	}

	if !resp.GetSuccess() {
		return "", fmt.Errorf("assignment failed for %s", id)
	}

	droneID := resp.GetDroneId()
	if droneID == "" {
		return "", fmt.Errorf("no drone assigned for %s", id)
	}

	return droneID, nil
}
