package dispatch

import (
	"context"
	"fmt"
	pbCommon "hive/gen/common"
	pb "hive/gen/dispatch"
	"hive/services/order/internal/domain/assignment"
	"hive/services/order/internal/domain/shared"
)

type DispatchClient struct {
	client pb.DispatchServiceClient
}

func NewDispatchClient(client pb.DispatchServiceClient) *DispatchClient {
	return &DispatchClient{client: client}
}

func (c *DispatchClient) AssignDrone(
	ctx context.Context,
	orderID string,
	deliveryLocation *shared.Location,
) (*assignment.AssignmentInfo, error) {
	req := &pb.AssignDroneRequest{
		OrderId: orderID,
		DeliveryLocation: &pbCommon.Location{
			Lat: deliveryLocation.Lat,
			Lon: deliveryLocation.Lon,
		},
	}
	resp, err := c.client.AssignDrone(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to assign drone: %w", err)
	}

	if resp.GetDroneId() == "" {
		return nil, fmt.Errorf("no drone assigned for %s", orderID)
	}

	return &assignment.AssignmentInfo{
		DroneID:    resp.GetDroneId(),
		EtaSeconds: resp.GetEtaSeconds(),
	}, nil
}
