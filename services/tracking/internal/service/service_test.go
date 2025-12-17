package service

import (
	"context"
	"hive/pkg/logger"
	"testing"

	pbCommon "hive/gen/common"
	pbTelemetry "hive/gen/telemetry"
	pb "hive/gen/tracking"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/domain/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type repoStub struct {
	GetNearestFn  func(context.Context, shared.Location, float64, float64) (*drone.DroneNearest, error)
	GetByIDFn     func(context.Context, string) (*drone.Drone, error)
	SetStatusFn   func(context.Context, string, drone.DroneStatus) error
	UpdateStateFn func(context.Context, drone.TelemetryData) error
	ListFn        func(context.Context, int64, int64) ([]*drone.Drone, error)
}

func (r *repoStub) GetNearest(ctx context.Context, loc shared.Location, minBattery, radius float64) (*drone.DroneNearest, error) {
	return r.GetNearestFn(ctx, loc, minBattery, radius)
}

func (r *repoStub) GetByID(ctx context.Context, id string) (*drone.Drone, error) {
	return r.GetByIDFn(ctx, id)
}

func (r *repoStub) SetStatus(ctx context.Context, id string, st drone.DroneStatus) error {
	return r.SetStatusFn(ctx, id, st)
}

func (r *repoStub) UpdateState(ctx context.Context, tm drone.TelemetryData) error {
	return r.UpdateStateFn(ctx, tm)
}

func (r *repoStub) List(ctx context.Context, offset, limit int64) ([]*drone.Drone, error) {
	if r.ListFn != nil {
		return r.ListFn(ctx, offset, limit)
	}
	return []*drone.Drone{}, nil
}

func TestTrackingService_FindNearest_InvalidArgument(t *testing.T) {
	repo := &repoStub{}
	svc := New(repo, lg)

	_, err := svc.FindNearest(context.Background(), &pb.FindNearestRequest{
		StoreLocation: nil,
		MinBattery:    10,
		RadiusMeters:  1000,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestTrackingService_FindNearest_NotFoundReturnsFoundFalse(t *testing.T) {
	repo := &repoStub{
		GetNearestFn: func(ctx context.Context, loc shared.Location, min, radius float64) (*drone.DroneNearest, error) {
			return nil, ErrDroneNotFound
		},
	}
	svc := New(repo, lg)

	resp, err := svc.FindNearest(context.Background(), &pb.FindNearestRequest{
		StoreLocation: &pbCommon.Location{Lat: 55, Lon: 37},
		MinBattery:    10,
		RadiusMeters:  1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetFound() {
		t.Fatalf("expected found=false")
	}
}

func TestTrackingService_SetStatus_Success(t *testing.T) {
	repo := &repoStub{
		SetStatusFn: func(ctx context.Context, id string, st drone.DroneStatus) error {
			return nil
		},
	}
	svc := New(repo, lg)

	resp, err := svc.SetStatus(context.Background(), &pb.SetStatusRequest{
		DroneId: "d1",
		Status:  pbTelemetry.DroneStatus_STATUS_BUSY,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatalf("expected success")
	}
}
