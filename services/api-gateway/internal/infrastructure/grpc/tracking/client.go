package tracking

import (
	"context"
	pbTracking "hive/gen/tracking"
	"hive/services/api-gateway/internal/domain/drone"
)

type TrackingClient struct {
	client pbTracking.TrackingServiceClient
}

func NewTrackingClient(client pbTracking.TrackingServiceClient) *TrackingClient {
	return &TrackingClient{
		client: client,
	}
}

func (c *TrackingClient) ListDrones(
	ctx context.Context,
	offset, limit int64,
) ([]*drone.Drone, error) {
	return []*drone.Drone{}, nil
}
