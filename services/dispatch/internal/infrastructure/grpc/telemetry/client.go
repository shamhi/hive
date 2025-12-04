package telemetry

import (
	"context"
	"fmt"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/dispatch/internal/domain"
)

type TelemetryAdapter struct {
	client pbTelemetry.TelemetryServiceClient
}

func NewTelemetryAdapter(client pbTelemetry.TelemetryServiceClient) *TelemetryAdapter {
	return &TelemetryAdapter{client: client}
}

func (a *TelemetryAdapter) SendCommand(
	ctx context.Context,
	droneID string,
	action domain.DroneAction,
	target domain.Location,
) error {
	var pbAction pbTelemetry.DroneAction
	switch action {
	case domain.DroneActionWait:
		pbAction = pbTelemetry.DroneAction_ACTION_WAIT
	case domain.DroneActionFlyTo:
		pbAction = pbTelemetry.DroneAction_ACTION_FLY_TO
	case domain.DroneActionPickupCargo:
		pbAction = pbTelemetry.DroneAction_ACTION_PICKUP_CARGO
	case domain.DroneActionDropCargo:
		pbAction = pbTelemetry.DroneAction_ACTION_DROP_CARGO
	case domain.DroneActionCharge:
		pbAction = pbTelemetry.DroneAction_ACTION_CHARGE
	}

	req := &pbTelemetry.SendCommandRequest{
		DroneId: droneID,
		Action:  pbAction,
		Target: &pbTelemetry.Location{
			Lat: target.Lat,
			Lon: target.Lon,
		},
	}

	resp, err := a.client.SendCommand(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send command to drone: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("drone telemetry service reported failure")
	}

	return nil
}
