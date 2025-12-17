package dispatch

import (
	"context"
	"fmt"
	pbDispatch "hive/gen/dispatch"
	"hive/services/api-gateway/internal/domain/assignment"
	"hive/services/api-gateway/internal/domain/mapping"
	"hive/services/api-gateway/internal/domain/shared"
)

type DispatchClient struct {
	client pbDispatch.DispatchServiceClient
}

func NewDispatchClient(client pbDispatch.DispatchServiceClient) *DispatchClient {
	return &DispatchClient{client: client}
}

func (c *DispatchClient) GetAssignment(
	ctx context.Context,
	droneID string,
) (*assignment.Assignment, error) {
	req := pbDispatch.GetAssignmentRequest{
		DroneId: droneID,
	}
	resp, err := c.client.GetAssignment(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment for %s drone: %w", droneID, err)
	}

	var tloc shared.Location
	if resp.GetTargetLocation() != nil {
		tloc = shared.Location{
			Lat: resp.GetTargetLocation().Lat,
			Lon: resp.GetTargetLocation().Lon,
		}
	}

	st, ok := mapping.AssignmentStatusFromProto(resp.GetStatus())
	if !ok {
		return nil, fmt.Errorf("invalid assignment status: %v", resp.GetStatus())
	}

	return &assignment.Assignment{
		ID:     resp.GetAssignmentId(),
		Status: st,
		Target: &tloc,
	}, nil
}
