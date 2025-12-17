package dispatch

import (
	"context"
	"fmt"
	pbDispatch "hive/gen/dispatch"
	"hive/services/api-gateway/internal/domain/assignment"
	"hive/services/api-gateway/internal/domain/mapping"
	"hive/services/api-gateway/internal/domain/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	req := &pbDispatch.GetAssignmentRequest{
		DroneId: droneID,
	}

	resp, err := c.client.GetAssignment(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	var tloc *shared.Location
	if tl := resp.GetTargetLocation(); tl != nil {
		tloc = &shared.Location{Lat: tl.GetLat(), Lon: tl.GetLon()}
	}

	st, ok := mapping.AssignmentStatusFromProto(resp.GetStatus())
	if !ok {
		return nil, fmt.Errorf("invalid assignment status: %v", resp.GetStatus())
	}

	return &assignment.Assignment{
		ID:     resp.GetAssignmentId(),
		Status: st,
		Target: tloc,
	}, nil
}
