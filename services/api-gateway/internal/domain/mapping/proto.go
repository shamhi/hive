package mapping

import (
	pbBase "hive/gen/base"
	pbCommon "hive/gen/common"
	pbOrder "hive/gen/order"
	pbStore "hive/gen/store"
	pbTelemetry "hive/gen/telemetry"
	"hive/services/api-gateway/internal/domain/base"
	"hive/services/api-gateway/internal/domain/drone"
	"hive/services/api-gateway/internal/domain/order"
	"hive/services/api-gateway/internal/domain/shared"
	"hive/services/api-gateway/internal/domain/store"
)

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

func BaseFromProto(b *pbBase.Base) (*base.Base, bool) {
	if b == nil {
		return nil, false
	}
	return &base.Base{
		ID:       b.BaseId,
		Name:     b.Name,
		Address:  b.Address,
		Location: *LocationFromProto(b.Location),
	}, true
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

func StoreFromProto(s *pbStore.Store) (*store.Store, bool) {
	if s == nil {
		return nil, false
	}
	return &store.Store{
		ID:       s.StoreId,
		Name:     s.Name,
		Address:  s.Address,
		Location: *LocationFromProto(s.Location),
	}, true
}

func DroneFromProto(d *pbTelemetry.Drone) (*drone.Drone, bool) {
	if d == nil {
		return nil, false
	}
	return &drone.Drone{
		ID:                  d.DroneId,
		Location:            *LocationFromProto(d.DroneLocation),
		Battery:             d.Battery,
		SpeedMps:            d.SpeedMps,
		ConsumptionPerMeter: d.ConsumptionPerMeter,
	}, true
}
