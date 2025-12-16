package mapping

import (
	pbCommon "hive/gen/common"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/domain/shared"
)

func LocationToProto(loc *shared.Location) *pbCommon.Location {
	if loc == nil {
		return nil
	}
	return &pbCommon.Location{
		Lat: loc.Lat,
		Lon: loc.Lon,
	}
}

func DroneStatusFromProto(status pbTelemetry.DroneStatus) (drone.DroneStatus, bool) {
	switch status {
	case pbTelemetry.DroneStatus_STATUS_FREE:
		return drone.DroneStatusFree, true
	case pbTelemetry.DroneStatus_STATUS_BUSY:
		return drone.DroneStatusBusy, true
	case pbTelemetry.DroneStatus_STATUS_CHARGING:
		return drone.DroneStatusCharging, true
	default:
		return "", false
	}
}

func DroneToProto(d *drone.Drone) *pbTelemetry.Drone {
	if d == nil {
		return nil
	}
	return &pbTelemetry.Drone{
		DroneId:             d.ID,
		DroneLocation:       LocationToProto(&d.Location),
		Battery:             d.Battery,
		SpeedMps:            d.SpeedMps,
		ConsumptionPerMeter: d.ConsumptionPerMeter,
	}
}
