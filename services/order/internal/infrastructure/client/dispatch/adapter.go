package dispatch

import (
	"context"
	pb "hive/gen/dispatch"
	"hive/services/order/internal/domain"
)

type DispatchAdapter struct {
	client pb.DispatchServiceClient
}

func NewDispatchAdapter(client pb.DispatchServiceClient) *DispatchAdapter {
	return &DispatchAdapter{client: client}
}

func (a *DispatchAdapter) AssignDrone(ctx context.Context, id string, loc domain.Location) error {
	req := &pb.AssignDroneRequest{
		OrderId: id,
		DeliveryLocation: &pb.Location{
			Lat: loc.Lat,
			Lon: loc.Lon,
		},
	}
	if _, err := a.client.AssignDrone(ctx, req); err != nil {
		return err
	}

	return nil
}
