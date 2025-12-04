package telemetry

import (
	"context"
	"fmt"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/dispatch/internal/domain/drone"
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
	var pbAction pbTelemetry.DroneAction
	switch action {
	case drone.DroneActionWait:
		pbAction = pbTelemetry.DroneAction_ACTION_WAIT
	case drone.DroneActionFlyTo:
		pbAction = pbTelemetry.DroneAction_ACTION_FLY_TO
	case drone.DroneActionPickupCargo:
		pbAction = pbTelemetry.DroneAction_ACTION_PICKUP_CARGO
	case drone.DroneActionDropCargo:
		pbAction = pbTelemetry.DroneAction_ACTION_DROP_CARGO
	case drone.DroneActionCharge:
		pbAction = pbTelemetry.DroneAction_ACTION_CHARGE
	default:
		pbAction = pbTelemetry.DroneAction_ACTION_NONE
	}

	var targetPb *pbTelemetry.Location
	var targetType pbTelemetry.TargetType
	if target != nil {
		targetPb = &pbTelemetry.Location{
			Lat: target.Location.Lat,
			Lon: target.Location.Lon,
		}

		switch target.Type {
		case drone.TargetTypePoint:
			targetType = pbTelemetry.TargetType_TARGET_POINT
		case drone.TargetTypeStore:
			targetType = pbTelemetry.TargetType_TARGET_STORE
		case drone.TargetTypeClient:
			targetType = pbTelemetry.TargetType_TARGET_CLIENT
		case drone.TargetTypeBase:
			targetType = pbTelemetry.TargetType_TARGET_BASE
		default:
			targetType = pbTelemetry.TargetType_TARGET_NONE
		}
	}

	req := &pbTelemetry.SendCommandRequest{
		DroneId: droneID,
		Action:  pbAction,
		Target:  targetPb,
		Type:    targetType,
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
