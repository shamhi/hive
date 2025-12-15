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

func LocationFromProto(loc *pbCommon.Location) *shared.Location {
	if loc == nil {
		return nil
	}
	return &shared.Location{
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

func OrderStatusFromProto(status pbOrder.OrderStatus) (order.OrderStatus, bool) {
	switch status {
	case pbOrder.OrderStatus_CREATED:
		return order.OrderStatusCreated, true
	case pbOrder.OrderStatus_PENDING:
		return order.OrderStatusPending, true
	case pbOrder.OrderStatus_ASSIGNED:
		return order.OrderStatusAssigned, true
	case pbOrder.OrderStatus_COMPLETED:
		return order.OrderStatusCompleted, true
	case pbOrder.OrderStatus_FAILED:
		return order.OrderStatusFailed, true
	default:
		return "", false
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

func DroneActionFromProto(action pbTelemetry.DroneAction) (drone.DroneAction, bool) {
	switch action {
	case pbTelemetry.DroneAction_ACTION_WAIT:
		return drone.DroneActionWait, true
	case pbTelemetry.DroneAction_ACTION_FLY_TO:
		return drone.DroneActionFlyTo, true
	case pbTelemetry.DroneAction_ACTION_PICKUP_CARGO:
		return drone.DroneActionPickupCargo, true
	case pbTelemetry.DroneAction_ACTION_DROP_CARGO:
		return drone.DroneActionDropCargo, true
	case pbTelemetry.DroneAction_ACTION_CHARGE:
		return drone.DroneActionCharge, true
	default:
		return "", false
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

func DroneTargetTypeFromProto(targetType pbTelemetry.TargetType) (drone.TargetType, bool) {
	switch targetType {
	case pbTelemetry.TargetType_TARGET_NONE:
		return drone.TargetTypeNone, true
	case pbTelemetry.TargetType_TARGET_POINT:
		return drone.TargetTypePoint, true
	case pbTelemetry.TargetType_TARGET_STORE:
		return drone.TargetTypeStore, true
	case pbTelemetry.TargetType_TARGET_CLIENT:
		return drone.TargetTypeClient, true
	case pbTelemetry.TargetType_TARGET_BASE:
		return drone.TargetTypeBase, true
	default:
		return "", false
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

func DroneEventFromProto(event pbTelemetry.DroneEvent) (drone.DroneEvent, bool) {
	switch event {
	case pbTelemetry.DroneEvent_EVENT_ARRIVED_AT_STORE:
		return drone.DroneEventArrivedAtStore, true
	case pbTelemetry.DroneEvent_EVENT_PICKED_UP_CARGO:
		return drone.DroneEventPickedUpCargo, true
	case pbTelemetry.DroneEvent_EVENT_ARRIVED_AT_CLIENT:
		return drone.DroneEventArrivedAtClient, true
	case pbTelemetry.DroneEvent_EVENT_DROPPED_CARGO:
		return drone.DroneEventDroppedCargo, true
	case pbTelemetry.DroneEvent_EVENT_ARRIVED_AT_BASE:
		return drone.DroneEventArrivedAtBase, true
	case pbTelemetry.DroneEvent_EVENT_FULLY_CHARGED:
		return drone.DroneEventFullyCharged, true
	default:
		return "", false
	}
}
