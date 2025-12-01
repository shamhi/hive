package grpc

import (
	"context"
	"errors"
	pb "hive/gen/order"
	"hive/services/order/internal/domain"
	"hive/services/order/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	orderService *service.OrderService
}

func NewServer(svc *service.OrderService) *Server {
	return &Server{orderService: svc}
}

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}
	if len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items list cannot be empty")
	}
	if req.GetDeliveryLocation() == nil {
		return nil, status.Error(codes.InvalidArgument, "delivery location is required")
	}

	orderID := uuid.NewString()

	loc := domain.Location{
		Lat: req.GetDeliveryLocation().GetLat(),
		Lon: req.GetDeliveryLocation().GetLon(),
	}

	droneID, etaSeconds, err := s.orderService.CreateOrder(
		ctx,
		orderID,
		req.GetUserId(),
		req.GetItems(),
		loc,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateOrderResponse{
		OrderId:    order.ID,
		Status:     toProtoStatus(order.Status),
		EtaSeconds: etaSeconds,
		DroneId:    droneID,
	}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	order, err := s.orderService.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetOrderResponse{
		OrderId:   order.ID,
		Status:    toProtoStatus(order.Status),
		DroneId:   order.DroneID,
		CreatedAt: order.CreatedAt.Unix(),
		UpdatedAt: order.UpdatedAt.Unix(),
	}, nil
}

func (s *Server) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.UpdateStatusResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	newStatus, ok := toDomainStatus(req.GetStatus())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid order status")
	}

	if err := s.orderService.UpdateStatus(ctx, req.GetOrderId(), newStatus); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateStatusResponse{Success: true}, nil
}
