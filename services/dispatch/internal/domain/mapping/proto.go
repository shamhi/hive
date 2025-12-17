package mapping

import (
	pbCommon "hive/gen/common"
	pb "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/order"
	"hive/services/dispatch/internal/domain/shared"
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

func AssignmentStatusToProto(status assignment.AssignmentStatus) pb.AssignmentStatus {
	switch status {
	case assignment.AssignmentStatusCreated:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_CREATED
	case assignment.AssignmentStatusAssigned:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_ASSIGNED
	case assignment.AssignmentStatusFlyingToStore:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_FLYING_TO_STORE
	case assignment.AssignmentStatusAtStore:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_AT_STORE
	case assignment.AssignmentStatusPickedUpCargo:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_PICKED_UP_CARGO
	case assignment.AssignmentStatusFlyingToClient:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_FLYING_TO_CLIENT
	case assignment.AssignmentStatusAtClient:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_AT_CLIENT
	case assignment.AssignmentStatusDroppedCargo:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_DROPPED_CARGO
	case assignment.AssignmentStatusReturningBase:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_RETURNING_BASE
	case assignment.AssignmentStatusCompleted:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED
	case assignment.AssignmentStatusFailed:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_FAILED
	default:
		return pb.AssignmentStatus_ASSIGNMENT_STATUS_UNKNOWN
	}
}

func OrderStatusToProto(status order.OrderStatus) pbOrder.OrderStatus {
	switch status {
	case order.OrderStatusCreated:
		return pbOrder.OrderStatus_CREATED
	case order.OrderStatusPending:
		return pbOrder.OrderStatus_PENDING
	case order.OrderStatusAssigned:
		return pbOrder.OrderStatus_ASSIGNED
	case order.OrderStatusCompleted:
		return pbOrder.OrderStatus_COMPLETED
	case order.OrderStatusFailed:
		return pbOrder.OrderStatus_FAILED
	default:
		return pbOrder.OrderStatus_UNKNOWN
	}
}

func DroneActionToProto(action drone.DroneAction) pbTelemetry.DroneAction {
	switch action {
	case drone.DroneActionWait:
		return pbTelemetry.DroneAction_ACTION_WAIT
	case drone.DroneActionFlyTo:
		return pbTelemetry.DroneAction_ACTION_FLY_TO
	case drone.DroneActionPickupCargo:
		return pbTelemetry.DroneAction_ACTION_PICKUP_CARGO
	case drone.DroneActionDropCargo:
		return pbTelemetry.DroneAction_ACTION_DROP_CARGO
	case drone.DroneActionCharge:
		return pbTelemetry.DroneAction_ACTION_CHARGE
	default:
		return pbTelemetry.DroneAction_ACTION_NONE
	}
}

func DroneTargetTypeToProto(targetType drone.TargetType) pbTelemetry.TargetType {
	switch targetType {
	case drone.TargetTypePoint:
		return pbTelemetry.TargetType_TARGET_POINT
	case drone.TargetTypeStore:
		return pbTelemetry.TargetType_TARGET_STORE
	case drone.TargetTypeClient:
		return pbTelemetry.TargetType_TARGET_CLIENT
	case drone.TargetTypeBase:
		return pbTelemetry.TargetType_TARGET_BASE
	default:
		return pbTelemetry.TargetType_TARGET_NONE
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
