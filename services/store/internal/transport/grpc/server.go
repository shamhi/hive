package grpc

import (
	"context"
	"errors"
	pbCommon "hive/gen/common"
	pb "hive/gen/store"
	"hive/services/store/internal/domain/mapping"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedStoreServiceServer
	svc    *service.StoreService
	config *Config
}

func NewServer(svc *service.StoreService, cfg *Config) *Server {
	return &Server{svc: svc, config: cfg}
}

func (s *Server) CreateStore(ctx context.Context, req *pb.CreateStoreRequest) (*pb.CreateStoreResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "store name is required")
	}
	if req.GetLocation() == nil {
		return nil, status.Error(codes.InvalidArgument, "store location is required")
	}

	storeID, err := s.svc.CreateStore(
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

	return &pb.CreateStoreResponse{StoreId: storeID}, nil
}

func (s *Server) GetStoreLocation(ctx context.Context, req *pb.GetStoreLocationRequest) (*pb.GetStoreLocationResponse, error) {
	if req.GetStoreId() == "" {
		return nil, status.Error(codes.InvalidArgument, "store ID is required")
	}

	storeInfo, err := s.svc.GetLocation(ctx, req.GetStoreId())
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			return nil, status.Error(codes.NotFound, "store not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStoreLocationResponse{
		Name:    storeInfo.Name,
		Address: storeInfo.Address,
		Location: &pbCommon.Location{
			Lat: storeInfo.Location.Lat,
			Lon: storeInfo.Location.Lon,
		},
	}, nil
}

func (s *Server) ListStores(ctx context.Context, req *pb.ListStoresRequest) (*pb.ListStoresResponse, error) {
	if req.GetLimit() <= 0 {
		return &pb.ListStoresResponse{
			Stores: []*pb.Store{},
		}, nil
	}

	stores, err := s.svc.ListStores(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbStores := make([]*pb.Store, 0, len(stores))
	for _, st := range stores {
		pbStores = append(pbStores, mapping.StoreToProto(st))
	}

	return &pb.ListStoresResponse{
		Stores: pbStores,
	}, nil
}

func (s *Server) FindNearest(ctx context.Context, req *pb.FindNearestRequest) (*pb.FindNearestResponse, error) {
	if req.GetDeliveryLocation() == nil {
		return nil, status.Error(codes.InvalidArgument, "delivery location is required")
	}

	storeInfo, err := s.svc.FindNearest(
		ctx,
		shared.Location{
			Lat: req.GetDeliveryLocation().GetLat(),
			Lon: req.GetDeliveryLocation().GetLon(),
		},
		s.config.SearchRadius,
	)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			return &pb.FindNearestResponse{Found: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if storeInfo == nil {
		return &pb.FindNearestResponse{Found: false}, nil
	}

	return &pb.FindNearestResponse{
		StoreId:        storeInfo.ID,
		Found:          true,
		DistanceMeters: storeInfo.Distance,
	}, nil
}
