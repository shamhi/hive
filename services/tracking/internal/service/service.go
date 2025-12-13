package service

import (
	"context"
	commonGen "hive/gen/common"
	"hive/gen/telemetry"
	trackingGen "hive/gen/tracking"
	"hive/services/tracking/internal/models"
	"hive/services/tracking/internal/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	trackingGen.TrackingServiceServer

	repo repository.TrackingRepository
}

func New(repo repository.TrackingRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) FindNearest(ctx context.Context, req *trackingGen.FindNearestRequest) (*trackingGen.FindNearestResponse, error) {
	radius := req.RadiusMeters
	if radius <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "radius must be greater than zero")
	}

	location := models.Location{
		Lat: req.StoreLocation.Lat,
		Lon: req.StoreLocation.Lon,
	}

	result, err := s.repo.FindNearest(ctx, location, radius)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find nearest drone: %v", err)
	}

	if len(result) == 0 {
		return &trackingGen.FindNearestResponse{
			Found: false,
		}, nil
	}

	nearest := result[0]

	return &trackingGen.FindNearestResponse{
		DroneId:        nearest.Name,
		Found:          true,
		DistanceMeters: nearest.Dist,
	}, nil
}

func (s *Service) GetDroneLocation(ctx context.Context, req *trackingGen.GetDroneLocationRequest) (*trackingGen.GetDroneLocationResponse, error) {
	droneGeo, err := s.repo.GetGeopostion(ctx, req.DroneId)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get drone location: %v", err)
	}

	return &trackingGen.GetDroneLocationResponse{
		Location: &commonGen.Location{
			Lat: droneGeo.Lat,
			Lon: droneGeo.Lon,
		},
	}, nil
}

func (s *Service) SetStatus(ctx context.Context, req *trackingGen.SetStatusRequest) (*trackingGen.SetStatusResponse, error) {
	if req.DroneId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id must not be empty")
	}

	switch req.Status {
	case telemetry.DroneStatus_STATUS_UNKNOWN:
		return nil, status.Errorf(codes.InvalidArgument, "status must be a valid DroneStatus, not UNKNOWN")
	case telemetry.DroneStatus_STATUS_FREE,
		telemetry.DroneStatus_STATUS_BUSY,
		telemetry.DroneStatus_STATUS_CHARGING:

	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid drone status: %v", req.Status)
	}

	err := s.repo.SetStatus(ctx, req.DroneId, int32(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set drone status: %v", err)
	}

	return &trackingGen.SetStatusResponse{
		Success: true,
	}, nil
}
