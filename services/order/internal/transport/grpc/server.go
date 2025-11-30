package grpc

import (
	"context"
	pb "hive/gen/order"
	"hive/services/order/internal/domain"
	"hive/services/order/internal/service"

	"github.com/google/uuid"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	orderService *service.OrderService
}

func NewServer(svc *service.OrderService) *Server {
	return &Server{orderService: svc}
}

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	id := uuid.NewString()
	loc := domain.Location{
		Lat: req.DeliveryLocation.Lat,
		Lon: req.DeliveryLocation.Lon,
	}
	estimatedTime, err := s.orderService.CreateOrder(ctx, id, req.GetItems(), loc)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderResponse{
		OrderId:       id,
		Status:        pb.OrderStatus_PENDING,
		EstimatedTime: estimatedTime,
	}, nil
}
