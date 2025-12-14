package mapping

import (
	pbCommon "hive/gen/common"
	pb "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbDrone "hive/gen/telemetry"
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

func DroneActionToProto(action drone.DroneAction) pbDrone.DroneAction {
	switch action {
	case drone.DroneActionWait:
		return pbDrone.DroneAction_ACTION_WAIT
	case drone.DroneActionFlyTo:
		return pbDrone.DroneAction_ACTION_FLY_TO
	case drone.DroneActionPickupCargo:
		return pbDrone.DroneAction_ACTION_PICKUP_CARGO
	case drone.DroneActionDropCargo:
		return pbDrone.DroneAction_ACTION_DROP_CARGO
	case drone.DroneActionCharge:
		return pbDrone.DroneAction_ACTION_CHARGE
	default:
		return pbDrone.DroneAction_ACTION_NONE
	}
}

func DroneActionFromProto(action pbDrone.DroneAction) (drone.DroneAction, bool) {
	switch action {
	case pbDrone.DroneAction_ACTION_WAIT:
		return drone.DroneActionWait, true
	case pbDrone.DroneAction_ACTION_FLY_TO:
		return drone.DroneActionFlyTo, true
	case pbDrone.DroneAction_ACTION_PICKUP_CARGO:
		return drone.DroneActionPickupCargo, true
	case pbDrone.DroneAction_ACTION_DROP_CARGO:
		return drone.DroneActionDropCargo, true
	case pbDrone.DroneAction_ACTION_CHARGE:
		return drone.DroneActionCharge, true
	default:
		return "", false
	}
}

func DroneTargetTypeToProto(targetType drone.TargetType) pbDrone.TargetType {
	switch targetType {
	case drone.TargetTypePoint:
		return pbDrone.TargetType_TARGET_POINT
	case drone.TargetTypeStore:
		return pbDrone.TargetType_TARGET_STORE
	case drone.TargetTypeClient:
		return pbDrone.TargetType_TARGET_CLIENT
	case drone.TargetTypeBase:
		return pbDrone.TargetType_TARGET_BASE
	default:
		return pbDrone.TargetType_TARGET_NONE
	}
}

func DroneTargetTypeFromProto(targetType pbDrone.TargetType) (drone.TargetType, bool) {
	switch targetType {
	case pbDrone.TargetType_TARGET_POINT:
		return drone.TargetTypePoint, true
	case pbDrone.TargetType_TARGET_STORE:
		return drone.TargetTypeStore, true
	case pbDrone.TargetType_TARGET_CLIENT:
		return drone.TargetTypeClient, true
	case pbDrone.TargetType_TARGET_BASE:
		return drone.TargetTypeBase, true
	default:
		return "", false
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

func DroneEventFromProto(event pbDrone.DroneEvent) (drone.DroneEvent, bool) {
	switch event {
	case pbDrone.DroneEvent_EVENT_ARRIVED_AT_STORE:
		return drone.DroneEventArrivedAtStore, true
	case pbDrone.DroneEvent_EVENT_PICKED_UP_CARGO:
		return drone.DroneEventPickedUpCargo, true
	case pbDrone.DroneEvent_EVENT_ARRIVED_AT_CLIENT:
		return drone.DroneEventArrivedAtClient, true
	case pbDrone.DroneEvent_EVENT_DROPPED_CARGO:
		return drone.DroneEventDroppedCargo, true
	case pbDrone.DroneEvent_EVENT_ARRIVED_AT_BASE:
		return drone.DroneEventArrivedAtBase, true
	case pbDrone.DroneEvent_EVENT_FULLY_CHARGED:
		return drone.DroneEventFullyCharged, true
	default:
		return "", false
	}
}
