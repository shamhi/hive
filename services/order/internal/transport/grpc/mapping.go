package grpc

import (
	pb "hive/gen/order"
	"hive/services/order/internal/domain/order"
)

func toProtoStatus(s order.OrderStatus) pb.OrderStatus {
	switch s {
	case order.OrderStatusCreated:
		return pb.OrderStatus_CREATED
	case order.OrderStatusPending:
		return pb.OrderStatus_PENDING
	case order.OrderStatusAssigned:
		return pb.OrderStatus_ASSIGNED
	case order.OrderStatusCompleted:
		return pb.OrderStatus_COMPLETED
	case order.OrderStatusFailed:
		return pb.OrderStatus_FAILED
	default:
		return pb.OrderStatus_FAILED
	}
}

func toDomainStatus(s pb.OrderStatus) (order.OrderStatus, bool) {
	switch s {
	case pb.OrderStatus_CREATED:
		return order.OrderStatusCreated, true
	case pb.OrderStatus_PENDING:
		return order.OrderStatusPending, true
	case pb.OrderStatus_ASSIGNED:
		return order.OrderStatusAssigned, true
	case pb.OrderStatus_COMPLETED:
		return order.OrderStatusCompleted, true
	case pb.OrderStatus_FAILED:
		return order.OrderStatusFailed, true
	default:
		return "", false
	}
}
