package grpc

import (
	"context"
	pb "hive/gen/dispatch"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedDispatchServiceServer
	dispatchService *service.DispatchService
	config          *Config
}

func NewServer(svc *service.DispatchService, cfg *Config) *Server {
	return &Server{
		dispatchService: svc,
		config:          cfg,
	}
}

func (s *Server) AssignDrone(ctx context.Context, req *pb.AssignDroneRequest) (*pb.AssignDroneResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}
	if req.GetDeliveryLocation() == nil {
		return nil, status.Error(codes.InvalidArgument, "delivery location is required")
	}

	droneID, err := s.dispatchService.AssignDrone(
		ctx,
		req.GetOrderId(),
		&shared.Location{
			Lat: req.GetDeliveryLocation().GetLat(),
			Lon: req.GetDeliveryLocation().GetLon(),
		},
		s.config.MinDroneBattery,
		s.config.DroneSearchRadius,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AssignDroneResponse{DroneId: droneID}, nil
}
