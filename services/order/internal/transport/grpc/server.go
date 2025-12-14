package grpc

import (
	"context"
	"errors"
	pbCommon "hive/gen/common"
	pb "hive/gen/order"
	"hive/services/order/internal/domain/mapping"
	"hive/services/order/internal/domain/shared"
	"hive/services/order/internal/service"

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

	orderInfo, err := s.orderService.CreateOrder(
		ctx,
		req.GetUserId(),
		req.GetItems(),
		shared.Location{
			Lat: req.GetDeliveryLocation().GetLat(),
			Lon: req.GetDeliveryLocation().GetLon(),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateOrderResponse{
		OrderId:    orderInfo.ID,
		Status:     mapping.OrderStatusToProto(orderInfo.Status),
		DroneId:    orderInfo.DroneID,
		EtaSeconds: orderInfo.EtaSeconds,
	}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	o, err := s.orderService.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetOrderResponse{
		OrderId: o.ID,
		UserId:  o.UserID,
		DroneId: o.DroneID,
		Items:   o.Items,
		Status:  mapping.OrderStatusToProto(o.Status),
		DeliveryLocation: &pbCommon.Location{
			Lat: o.Location.Lat,
			Lon: o.Location.Lon,
		},
	}, nil
}

func (s *Server) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.UpdateStatusResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	newStatus, ok := mapping.OrderStatusFromProto(req.GetStatus())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid order status")
	}

	if err := s.orderService.UpdateStatus(ctx, req.GetOrderId(), newStatus); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateStatusResponse{Success: true}, nil
}
