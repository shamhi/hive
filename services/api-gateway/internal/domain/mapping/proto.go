package mapping

import (
	pbBase "hive/gen/base"
	pbCommon "hive/gen/common"
	pb "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbStore "hive/gen/store"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/api-gateway/internal/domain/assignment"
	"hive/services/api-gateway/internal/domain/base"
	"hive/services/api-gateway/internal/domain/drone"
	"hive/services/api-gateway/internal/domain/order"
	"hive/services/api-gateway/internal/domain/shared"
	"hive/services/api-gateway/internal/domain/store"
)

func LocationFromProto(loc *pbCommon.Location) (shared.Location, bool) {
	if loc == nil {
		return shared.Location{}, false
	}
	return shared.Location{
		Lat: loc.Lat,
		Lon: loc.Lon,
	}, true
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

func AssignmentStatusFromProto(status pb.AssignmentStatus) (assignment.AssignmentStatus, bool) {
	switch status {
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_CREATED:
		return assignment.AssignmentStatusCreated, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_ASSIGNED:
		return assignment.AssignmentStatusAssigned, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_FLYING_TO_STORE:
		return assignment.AssignmentStatusFlyingToStore, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_AT_STORE:
		return assignment.AssignmentStatusAtStore, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_PICKED_UP_CARGO:
		return assignment.AssignmentStatusPickedUpCargo, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_FLYING_TO_CLIENT:
		return assignment.AssignmentStatusFlyingToClient, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_AT_CLIENT:
		return assignment.AssignmentStatusAtClient, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_DROPPED_CARGO:
		return assignment.AssignmentStatusDroppedCargo, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_RETURNING_BASE:
		return assignment.AssignmentStatusReturningBase, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED:
		return assignment.AssignmentStatusCompleted, true
	case pb.AssignmentStatus_ASSIGNMENT_STATUS_FAILED:
		return assignment.AssignmentStatusFailed, true
	default:
		return "", false
	}
}

func BaseFromProto(b *pbBase.Base) (*base.Base, bool) {
	if b == nil {
		return nil, false
	}
	loc, ok := LocationFromProto(b.Location)
	if !ok {
		return nil, false
	}
	return &base.Base{
		ID:       b.BaseId,
		Name:     b.Name,
		Address:  b.Address,
		Location: loc,
	}, true
}

func StoreFromProto(s *pbStore.Store) (*store.Store, bool) {
	if s == nil {
		return nil, false
	}
	loc, ok := LocationFromProto(s.Location)
	if !ok {
		return nil, false
	}
	return &store.Store{
		ID:       s.StoreId,
		Name:     s.Name,
		Address:  s.Address,
		Location: loc,
	}, true
}

func DroneFromProto(d *pbTelemetry.Drone) (*drone.Drone, bool) {
	if d == nil {
		return nil, false
	}
	loc, ok := LocationFromProto(d.DroneLocation)
	if !ok {
		return nil, false
	}
	st, ok := DroneStatusFromProto(d.Status)
	if !ok {
		return nil, false
	}
	return &drone.Drone{
		ID:                  d.DroneId,
		Location:            loc,
		Battery:             d.Battery,
		SpeedMps:            d.SpeedMps,
		ConsumptionPerMeter: d.ConsumptionPerMeter,
		Status:              st,
		UpdatedAt:           d.UpdatedAt,
	}, true
}
