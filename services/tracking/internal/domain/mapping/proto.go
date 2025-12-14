package mapping

import (
	pbCommon "hive/gen/common"
	pbDrone "hive/gen/telemetry"
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

func DroneStatusToProto(status drone.DroneStatus) pbDrone.DroneStatus {
	switch status {
	case drone.DroneStatusFree:
		return pbDrone.DroneStatus_STATUS_FREE
	case drone.DroneStatusBusy:
		return pbDrone.DroneStatus_STATUS_BUSY
	case drone.DroneStatusCharging:
		return pbDrone.DroneStatus_STATUS_CHARGING
	default:
		return pbDrone.DroneStatus_STATUS_UNKNOWN
	}
}

func DroneStatusFromProto(status pbDrone.DroneStatus) (drone.DroneStatus, bool) {
	switch status {
	case pbDrone.DroneStatus_STATUS_FREE:
		return drone.DroneStatusFree, true
	case pbDrone.DroneStatus_STATUS_BUSY:
		return drone.DroneStatusBusy, true
	case pbDrone.DroneStatus_STATUS_CHARGING:
		return drone.DroneStatusCharging, true
	default:
		return "", false
	}
}
