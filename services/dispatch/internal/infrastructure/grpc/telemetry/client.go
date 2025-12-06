package telemetry

import (
	"context"
	"fmt"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/mapping"
)

type TelemetryClient struct {
	client pbTelemetry.TelemetryServiceClient
}

func NewTelemetryClient(client pbTelemetry.TelemetryServiceClient) *TelemetryClient {
	return &TelemetryClient{client: client}
}

func (c *TelemetryClient) SendCommand(
	ctx context.Context,
	droneID string,
	action drone.DroneAction,
	target *drone.Target,
) error {
	req := &pbTelemetry.SendCommandRequest{
		DroneId: droneID,
		Action:  mapping.DroneActionToProto(action),
		Type:    pbTelemetry.TargetType_TARGET_NONE,
	}
	if target != nil && target.Location != nil {
		req.Target = mapping.LocationToProto(target.Location)
		req.Type = mapping.DroneTargetTypeToProto(target.Type)
	} else if action == drone.DroneActionFlyTo {
		return fmt.Errorf("target is required for action %s", action)
	}

	resp, err := c.client.SendCommand(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send command to drone: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("drone telemetry service reported failure")
	}

	return nil
}
