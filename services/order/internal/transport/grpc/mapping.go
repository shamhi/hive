package grpc

import (
	pb "hive/gen/order"
	"hive/services/order/internal/domain"
)

func toProtoStatus(s domain.OrderStatus) pb.OrderStatus {
	switch s {
	case domain.OrderStatusCreated:
		return pb.OrderStatus_CREATED
	case domain.OrderStatusPending:
		return pb.OrderStatus_PENDING
	case domain.OrderStatusAssigned:
		return pb.OrderStatus_ASSIGNED
	case domain.OrderStatusCompleted:
		return pb.OrderStatus_COMPLETED
	case domain.OrderStatusFailed:
		return pb.OrderStatus_FAILED
	default:
		return pb.OrderStatus_FAILED
	}
}

func toDomainStatus(s pb.OrderStatus) (domain.OrderStatus, bool) {
	switch s {
	case pb.OrderStatus_CREATED:
		return domain.OrderStatusCreated, true
	case pb.OrderStatus_PENDING:
		return domain.OrderStatusPending, true
	case pb.OrderStatus_ASSIGNED:
		return domain.OrderStatusAssigned, true
	case pb.OrderStatus_COMPLETED:
		return domain.OrderStatusCompleted, true
	case pb.OrderStatus_FAILED:
		return domain.OrderStatusFailed, true
	default:
		return "", false
	}
}
