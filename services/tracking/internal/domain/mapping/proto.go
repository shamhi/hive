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

func LocationFromProto(loc *pbCommon.Location) *shared.Location {
	if loc == nil {
		return nil
	}
	return &shared.Location{
		Lat: loc.Lat,
		Lon: loc.Lon,
	}
}

func DroneStatusToProto(status drone.DroneStatus) pbTelemetry.DroneStatus {
	switch status {
	case drone.DroneStatusFree:
		return pbTelemetry.DroneStatus_STATUS_FREE
	case drone.DroneStatusBusy:
		return pbTelemetry.DroneStatus_STATUS_BUSY
	case drone.DroneStatusCharging:
		return pbTelemetry.DroneStatus_STATUS_CHARGING
	default:
		return pbTelemetry.DroneStatus_STATUS_UNKNOWN
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
