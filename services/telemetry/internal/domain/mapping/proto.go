package mapping

import (
	pbCommon "hive/gen/common"
	pb "hive/gen/telemetry"
	"hive/services/telemetry/internal/domain/drone"
	"hive/services/telemetry/internal/domain/shared"
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

func DroneActionFromProto(action pb.DroneAction) (drone.DroneAction, bool) {
	switch action {
	case pb.DroneAction_ACTION_WAIT:
		return drone.DroneActionWait, true
	case pb.DroneAction_ACTION_FLY_TO:
		return drone.DroneActionFlyTo, true
	case pb.DroneAction_ACTION_PICKUP_CARGO:
		return drone.DroneActionPickupCargo, true
	case pb.DroneAction_ACTION_DROP_CARGO:
		return drone.DroneActionDropCargo, true
	case pb.DroneAction_ACTION_CHARGE:
		return drone.DroneActionCharge, true
	default:
		return "", false
	}
}

func DroneActionToProto(action drone.DroneAction) pb.DroneAction {
	switch action {
	case drone.DroneActionWait:
		return pb.DroneAction_ACTION_WAIT
	case drone.DroneActionFlyTo:
		return pb.DroneAction_ACTION_FLY_TO
	case drone.DroneActionPickupCargo:
		return pb.DroneAction_ACTION_PICKUP_CARGO
	case drone.DroneActionDropCargo:
		return pb.DroneAction_ACTION_DROP_CARGO
	case drone.DroneActionCharge:
		return pb.DroneAction_ACTION_CHARGE
	default:
		return pb.DroneAction_ACTION_NONE
	}
}

func DroneTargetTypeToProto(targetType drone.TargetType) pb.TargetType {
	switch targetType {
	case drone.TargetTypePoint:
		return pb.TargetType_TARGET_POINT
	case drone.TargetTypeStore:
		return pb.TargetType_TARGET_STORE
	case drone.TargetTypeClient:
		return pb.TargetType_TARGET_CLIENT
	case drone.TargetTypeBase:
		return pb.TargetType_TARGET_BASE
	default:
		return pb.TargetType_TARGET_NONE
	}
}

func DroneTargetTypeFromProto(targetType pb.TargetType) (drone.TargetType, bool) {
	switch targetType {
	case pb.TargetType_TARGET_POINT:
		return drone.TargetTypePoint, true
	case pb.TargetType_TARGET_STORE:
		return drone.TargetTypeStore, true
	case pb.TargetType_TARGET_CLIENT:
		return drone.TargetTypeClient, true
	case pb.TargetType_TARGET_BASE:
		return drone.TargetTypeBase, true
	default:
		return "", false
	}
}

func DroneStatusToProto(status drone.DroneStatus) pb.DroneStatus {
	switch status {
	case drone.DroneStatusFree:
		return pb.DroneStatus_STATUS_FREE
	case drone.DroneStatusBusy:
		return pb.DroneStatus_STATUS_BUSY
	case drone.DroneStatusCharging:
		return pb.DroneStatus_STATUS_CHARGING
	default:
		return pb.DroneStatus_STATUS_UNKNOWN
	}
}

func DroneStatusFromProto(status pb.DroneStatus) (drone.DroneStatus, bool) {
	switch status {
	case pb.DroneStatus_STATUS_FREE:
		return drone.DroneStatusFree, true
	case pb.DroneStatus_STATUS_BUSY:
		return drone.DroneStatusBusy, true
	case pb.DroneStatus_STATUS_CHARGING:
		return drone.DroneStatusCharging, true
	default:
		return "", false
	}
}

func DroneEventFromProto(event pb.DroneEvent) (drone.DroneEvent, bool) {
	switch event {
	case pb.DroneEvent_EVENT_ARRIVED_AT_STORE:
		return drone.DroneEventArrivedAtStore, true
	case pb.DroneEvent_EVENT_PICKED_UP_CARGO:
		return drone.DroneEventPickedUpCargo, true
	case pb.DroneEvent_EVENT_ARRIVED_AT_CLIENT:
		return drone.DroneEventArrivedAtClient, true
	case pb.DroneEvent_EVENT_DROPPED_CARGO:
		return drone.DroneEventDroppedCargo, true
	case pb.DroneEvent_EVENT_ARRIVED_AT_BASE:
		return drone.DroneEventArrivedAtBase, true
	case pb.DroneEvent_EVENT_FULLY_CHARGED:
		return drone.DroneEventFullyCharged, true
	default:
		return "", false
	}
}
