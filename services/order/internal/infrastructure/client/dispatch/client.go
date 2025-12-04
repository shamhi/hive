package dispatch

import (
	"context"
	"fmt"
	pb "hive/gen/dispatch"
	"hive/services/order/internal/domain"
)

type DispatchClient struct {
	client pb.DispatchServiceClient
}

func NewDispatchAdapter(client pb.DispatchServiceClient) *DispatchClient {
	return &DispatchClient{client: client}
}

func (c *DispatchClient) AssignDrone(ctx context.Context, id string, loc domain.Location) (string, error) {
	req := &pb.AssignDroneRequest{
		OrderId: id,
		DeliveryLocation: &pb.Location{
			Lat: loc.Lat,
			Lon: loc.Lon,
		},
	}
	resp, err := c.client.AssignDrone(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to assign drone: %w", err)
	}

	droneID := resp.GetDroneId()
	if droneID == "" {
		return "", fmt.Errorf("no drone assigned for %s", id)
	}

	return droneID, nil
}
