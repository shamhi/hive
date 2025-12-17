package tracking

import (
	"context"
	"fmt"

	pbTracking "hive/gen/tracking"
	"hive/services/api-gateway/internal/domain/drone"
	"hive/services/api-gateway/internal/domain/mapping"
)

type TrackingClient struct {
	client pbTracking.TrackingServiceClient
}

func NewTrackingClient(client pbTracking.TrackingServiceClient) *TrackingClient {
	return &TrackingClient{client: client}
}

func (c *TrackingClient) ListDrones(ctx context.Context, offset, limit int64) ([]*drone.Drone, error) {
	req := &pbTracking.ListDronesRequest{
		Offset: offset,
		Limit:  limit,
	}

	resp, err := c.client.ListDrones(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list drones: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("received nil response when listing drones")
	}

	pbDrones := resp.GetDrones()
	out := make([]*drone.Drone, 0, len(pbDrones))
	for _, pbD := range pbDrones {
		d, ok := mapping.DroneFromProto(pbD)
		if !ok || d == nil {
			continue
		}
		out = append(out, d)
	}

	return out, nil
}
