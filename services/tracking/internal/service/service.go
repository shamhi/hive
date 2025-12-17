package service

import (
	"context"
	"errors"
	pbCommon "hive/gen/common"
	pbTelemetry "hive/gen/telemetry"
	pb "hive/gen/tracking"
	"hive/pkg/logger"
	"hive/services/tracking/internal/domain/mapping"
	"hive/services/tracking/internal/domain/shared"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.TrackingServiceServer

	repo DroneRepository
	lg   logger.Logger
}

func New(
	repo DroneRepository,
	lg logger.Logger,
) *Service {
	return &Service{
		repo: repo,
		lg:   lg,
	}
}

func (s *Service) FindNearest(ctx context.Context, req *pb.FindNearestRequest) (*pb.FindNearestResponse, error) {
	lg := s.lg.With(
		zap.String("component", "tracking_service"),
		zap.String("op", "FindNearest"),
	)

	start := time.Now()

	if req.GetStoreLocation() == nil {
		lg.Warn(ctx, "validation failed: store_location is nil")
		return nil, status.Errorf(codes.InvalidArgument, "store_location must be provided")
	}
	if req.GetRadiusMeters() <= 0 {
		lg.Warn(ctx, "validation failed: radius_meters <= 0", zap.Float64("radius_meters", req.GetRadiusMeters()))
		return nil, status.Errorf(codes.InvalidArgument, "radius_meters must be greater than zero")
	}
	if req.GetMinBattery() < 0 || req.GetMinBattery() > 100 {
		lg.Warn(ctx, "validation failed: min_battery out of range", zap.Float64("min_battery", req.GetMinBattery()))
		return nil, status.Errorf(codes.InvalidArgument, "min_battery must be between 0 and 100")
	}

	loc := shared.Location{
		Lat: req.StoreLocation.Lat,
		Lon: req.StoreLocation.Lon,
	}

	lg.Info(ctx, "find nearest started",
		zap.Float64("store_lat", loc.Lat),
		zap.Float64("store_lon", loc.Lon),
		zap.Float64("radius_meters", req.GetRadiusMeters()),
		zap.Float64("min_battery", req.GetMinBattery()),
	)

	droneNearest, err := s.repo.GetNearest(ctx, loc, req.GetRadiusMeters(), req.GetMinBattery())
	if err != nil {
		if errors.Is(err, ErrDroneNotFound) {
			lg.Info(ctx, "no drone found",
				zap.Duration("duration", time.Since(start)),
			)
			return &pb.FindNearestResponse{Found: false}, nil
		}
		lg.Error(ctx, "failed to find nearest drone", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, status.Errorf(codes.Internal, "failed to find nearest drone: %v", err)
	}

	lg.Info(ctx, "nearest drone found",
		zap.String("drone_id", droneNearest.ID),
		zap.Float64("distance_meters", droneNearest.Distance),
		zap.Duration("duration", time.Since(start)),
	)

	return &pb.FindNearestResponse{
		Found:          true,
		DroneId:        droneNearest.ID,
		DistanceMeters: droneNearest.Distance,
	}, nil
}

func (s *Service) GetDroneLocation(ctx context.Context, req *pb.GetDroneLocationRequest) (*pb.GetDroneLocationResponse, error) {
	lg := s.lg.With(
		zap.String("component", "tracking_service"),
		zap.String("op", "GetDroneLocation"),
		zap.String("drone_id", req.GetDroneId()),
	)

	start := time.Now()

	if req.GetDroneId() == "" {
		lg.Warn(ctx, "validation failed: drone_id is empty")
		return nil, status.Errorf(codes.InvalidArgument, "drone_id must not be empty")
	}

	lg.Info(ctx, "get drone location started")
	d, err := s.repo.GetByID(ctx, req.GetDroneId())
	if err != nil {
		if errors.Is(err, ErrDroneNotFound) {
			lg.Info(ctx, "drone not found", zap.Duration("duration", time.Since(start)))
			return nil, status.Errorf(codes.NotFound, "failed to get drone location: %v", err)
		}
		lg.Error(ctx, "failed to get drone location", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, status.Errorf(codes.NotFound, "failed to get drone location: %v", err)
	}

	lg.Info(ctx, "get drone location completed",
		zap.Float64("lat", d.Location.Lat),
		zap.Float64("lon", d.Location.Lon),
		zap.Float64("battery", d.Battery),
		zap.Float64("speed_mps", d.SpeedMps),
		zap.Float64("consumption_per_meter", d.ConsumptionPerMeter),
		zap.Duration("duration", time.Since(start)),
	)

	return &pb.GetDroneLocationResponse{
		Location: &pbCommon.Location{
			Lat: d.Location.Lat,
			Lon: d.Location.Lon,
		},
		Battery:             d.Battery,
		SpeedMps:            d.SpeedMps,
		ConsumptionPerMeter: d.ConsumptionPerMeter,
		Status:              mapping.DroneStatusToProto(d.Status),
	}, nil
}

func (s *Service) SetStatus(ctx context.Context, req *pb.SetStatusRequest) (*pb.SetStatusResponse, error) {
	lg := s.lg.With(
		zap.String("component", "tracking_service"),
		zap.String("op", "SetStatus"),
		zap.String("drone_id", req.GetDroneId()),
		zap.Int32("status_proto", int32(req.GetStatus())),
	)

	start := time.Now()

	if req.GetDroneId() == "" {
		lg.Warn(ctx, "validation failed: drone_id is empty")
		return nil, status.Errorf(codes.InvalidArgument, "drone_id must not be empty")
	}

	st, ok := mapping.DroneStatusFromProto(req.GetStatus())
	if !ok {
		lg.Warn(ctx, "validation failed: invalid drone status", zap.Int32("status_proto", int32(req.GetStatus())))
		return nil, status.Errorf(codes.InvalidArgument, "invalid drone status: %v", req.GetStatus())
	}

	lg.Info(ctx, "set status started", zap.String("status", string(st)))
	if err := s.repo.SetStatus(ctx, req.DroneId, st); err != nil {
		lg.Error(ctx, "failed to set drone status", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, status.Errorf(codes.Internal, "failed to set drone status: %v", err)
	}

	lg.Info(ctx, "set status completed", zap.String("status", string(st)), zap.Duration("duration", time.Since(start)))

	return &pb.SetStatusResponse{Success: true}, nil
}

func (s *Service) ListDrones(ctx context.Context, req *pb.ListDronesRequest) (*pb.ListDronesResponse, error) {
	lg := s.lg.With(
		zap.String("component", "tracking_service"),
		zap.String("op", "ListDrones"),
		zap.Int64("offset", req.GetOffset()),
		zap.Int64("limit", req.GetLimit()),
	)

	start := time.Now()

	if req.GetLimit() <= 0 {
		lg.Info(ctx, "limit <= 0, returning empty list", zap.Duration("duration", time.Since(start)))
		return &pb.ListDronesResponse{Drones: []*pbTelemetry.Drone{}}, nil
	}

	drones, err := s.repo.List(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		lg.Error(ctx, "failed to list drones", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, status.Errorf(codes.Internal, "failed to list drones: %v", err)
	}

	pbDrones := make([]*pbTelemetry.Drone, 0, len(drones))
	for _, d := range drones {
		pbDrones = append(pbDrones, mapping.DroneToProto(d))
	}

	lg.Info(ctx, "list drones completed",
		zap.Int("count", len(pbDrones)),
		zap.Duration("duration", time.Since(start)),
	)

	return &pb.ListDronesResponse{
		Drones: pbDrones,
	}, nil
}
