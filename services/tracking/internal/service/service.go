package service

import (
	"context"
	"errors"
	pbCommon "hive/gen/common"
	pbTelemetry "hive/gen/telemetry"
	pb "hive/gen/tracking"
	"hive/services/tracking/internal/domain/mapping"
	"hive/services/tracking/internal/domain/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.TrackingServiceServer

	repo DroneRepository
}

func New(repo DroneRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) FindNearest(ctx context.Context, req *pb.FindNearestRequest) (*pb.FindNearestResponse, error) {
	if req.GetStoreLocation() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "store_location must be provided")
	}
	if req.GetRadiusMeters() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "radius_meters must be greater than zero")
	}
	if req.GetMinBattery() < 0 || req.GetMinBattery() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "min_battery must be between 0 and 100")
	}

	loc := shared.Location{
		Lat: req.StoreLocation.Lat,
		Lon: req.StoreLocation.Lon,
	}
	droneNearest, err := s.repo.GetNearest(ctx, loc, req.GetRadiusMeters(), req.GetMinBattery())
	if err != nil {
		if errors.Is(err, ErrDroneNotFound) {
			return &pb.FindNearestResponse{
				Found: false,
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to find nearest drone: %v", err)
	}

	return &pb.FindNearestResponse{
		Found:          true,
		DroneId:        droneNearest.ID,
		DistanceMeters: droneNearest.Distance,
	}, nil
}

func (s *Service) GetDroneLocation(ctx context.Context, req *pb.GetDroneLocationRequest) (*pb.GetDroneLocationResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id must not be empty")
	}

	d, err := s.repo.GetByID(ctx, req.GetDroneId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get drone location: %v", err)
	}

	return &pb.GetDroneLocationResponse{
		Location: &pbCommon.Location{
			Lat: d.Location.Lat,
			Lon: d.Location.Lon,
		},
		Battery:             d.Battery,
		SpeedMps:            d.SpeedMps,
		ConsumptionPerMeter: d.ConsumptionPerMeter,
	}, nil
}

func (s *Service) SetStatus(ctx context.Context, req *pb.SetStatusRequest) (*pb.SetStatusResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id must not be empty")
	}

	st, ok := mapping.DroneStatusFromProto(req.GetStatus())
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid drone status: %v", req.GetStatus())
	}

	err := s.repo.SetStatus(ctx, req.DroneId, st)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set drone status: %v", err)
	}

	return &pb.SetStatusResponse{
		Success: true,
	}, nil
}

func (s *Service) ListDrones(ctx context.Context, req *pb.ListDronesRequest) (*pb.ListDronesResponse, error) {
	if req.GetLimit() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "limit must be greater than zero")
	}

	drones, err := s.repo.List(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list drones: %v", err)
	}

	pbDrones := make([]*pbTelemetry.Drone, 0, len(drones))
	for _, d := range drones {
		pbDrones = append(pbDrones, mapping.DroneToProto(d))
	}

	return &pb.ListDronesResponse{
		Drones: pbDrones,
	}, nil
}
