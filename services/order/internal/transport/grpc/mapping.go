package grpc

import (
	pb "hive/gen/order"
	"hive/services/order/internal/domain"
)

func toProtoStatus(s domain.OrderStatus) pb.OrderStatus {
	switch s {
	case domain.CREATED:
		return pb.OrderStatus_CREATED
	case domain.PENDING:
		return pb.OrderStatus_PENDING
	case domain.ASSIGNED:
		return pb.OrderStatus_ASSIGNED
	case domain.COMPLETED:
		return pb.OrderStatus_COMPLETED
	case domain.FAILED:
		return pb.OrderStatus_FAILED
	default:
		return pb.OrderStatus_FAILED
	}
}

func toDomainStatus(s pb.OrderStatus) (domain.OrderStatus, bool) {
	switch s {
	case pb.OrderStatus_CREATED:
		return domain.CREATED, true
	case pb.OrderStatus_PENDING:
		return domain.PENDING, true
	case pb.OrderStatus_ASSIGNED:
		return domain.ASSIGNED, true
	case pb.OrderStatus_COMPLETED:
		return domain.COMPLETED, true
	case pb.OrderStatus_FAILED:
		return domain.FAILED, true
	default:
		return "", false
	}
}
