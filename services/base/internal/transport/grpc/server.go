package grpc

import (
	"context"
	"errors"
	pb "hive/gen/base"
	pbCommon "hive/gen/common"
	"hive/services/base/internal/domain/mapping"
	"hive/services/base/internal/domain/shared"
	"hive/services/base/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedBaseServiceServer
	svc    *service.BaseService
	config *Config
}

func NewServer(svc *service.BaseService, cfg *Config) *Server {
	return &Server{svc: svc, config: cfg}
}

func (s *Server) CreateBase(ctx context.Context, req *pb.CreateBaseRequest) (*pb.CreateBaseResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "base name is required")
	}
	if req.GetLocation() == nil {
		return nil, status.Error(codes.InvalidArgument, "base location is required")
	}

	baseID, err := s.svc.CreateBase(
		ctx,
		req.GetName(),
		req.GetAddress(),
		shared.Location{
			Lat: req.GetLocation().GetLat(),
			Lon: req.GetLocation().GetLon(),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateBaseResponse{BaseId: baseID}, nil
}

func (s *Server) GetBaseLocation(ctx context.Context, req *pb.GetBaseLocationRequest) (*pb.GetBaseLocationResponse, error) {
	if req.GetBaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "base ID is required")
	}

	baseInfo, err := s.svc.GetLocation(ctx, req.GetBaseId())
	if err != nil {
		if errors.Is(err, service.ErrBaseNotFound) {
			return nil, status.Error(codes.NotFound, "base not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetBaseLocationResponse{
		Name:    baseInfo.Name,
		Address: baseInfo.Address,
		Location: &pbCommon.Location{
			Lat: baseInfo.Location.Lat,
			Lon: baseInfo.Location.Lon,
		},
	}, nil
}

func (s *Server) ListBases(ctx context.Context, req *pb.ListBasesRequest) (*pb.ListBasesResponse, error) {
	if req.GetLimit() <= 0 {
		return &pb.ListBasesResponse{
			Bases: []*pb.Base{},
		}, nil
	}

	bases, err := s.svc.ListBases(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbBases := make([]*pb.Base, 0, len(bases))
	for _, st := range bases {
		pbBases = append(pbBases, mapping.BaseToProto(st))
	}

	return &pb.ListBasesResponse{
		Bases: pbBases,
	}, nil
}

func (s *Server) FindNearest(ctx context.Context, req *pb.FindNearestRequest) (*pb.FindNearestResponse, error) {
	if req.GetDroneLocation() == nil {
		return nil, status.Error(codes.InvalidArgument, "drone location is required")
	}

	baseInfo, err := s.svc.FindNearest(
		ctx,
		shared.Location{
			Lat: req.GetDroneLocation().GetLat(),
			Lon: req.GetDroneLocation().GetLon(),
		},
		s.config.SearchRadius,
	)
	if err != nil {
		if errors.Is(err, service.ErrBaseNotFound) {
			return &pb.FindNearestResponse{Found: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if baseInfo == nil {
		return &pb.FindNearestResponse{Found: false}, nil
	}

	return &pb.FindNearestResponse{
		BaseId:         baseInfo.ID,
		Found:          true,
		DistanceMeters: baseInfo.Distance,
	}, nil
}
