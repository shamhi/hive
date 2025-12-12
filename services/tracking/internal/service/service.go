package service

import (
	"context"
	trackingGen "hive/gen/tracking"
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

func (s *Service) FindNearest(context.Context, *trackingGen.FindNearestRequest) (*trackingGen.FindNearestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method not implemented")
}

func (s *Service) GetDroneLocation(ctx context.Context, req *trackingGen.GetDroneLocationRequest) (*trackingGen.GetDroneLocationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method not implemented")
}

func (s *Service) SetStatus(context.Context, *trackingGen.SetStatusRequest) (*trackingGen.SetStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method not implemented")
}
