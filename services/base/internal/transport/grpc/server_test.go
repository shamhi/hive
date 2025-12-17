package grpc

import (
	"context"
	"errors"
	"hive/pkg/logger"
	"testing"

	pb "hive/gen/base"
	pbCommon "hive/gen/common"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"
	basesvc "hive/services/base/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type baseRepoStub struct {
	GetByIDFn    func(context.Context, string) (*base.Base, error)
	GetNearestFn func(context.Context, shared.Location, float64) (*base.BaseNearest, error)
}

func (r *baseRepoStub) Save(context.Context, *base.Base) error { return nil }
func (r *baseRepoStub) List(context.Context, int64, int64) ([]*base.Base, error) {
	return []*base.Base{}, nil
}

func (r *baseRepoStub) GetByID(ctx context.Context, id string) (*base.Base, error) {
	if r.GetByIDFn != nil {
		return r.GetByIDFn(ctx, id)
	}
	return nil, basesvc.ErrBaseNotFound
}

func (r *baseRepoStub) GetNearest(ctx context.Context, loc shared.Location, radius float64) (*base.BaseNearest, error) {
	if r.GetNearestFn != nil {
		return r.GetNearestFn(ctx, loc, radius)
	}
	return nil, basesvc.ErrBaseNotFound
}

func TestServer_CreateBase_InvalidArgument(t *testing.T) {
	repo := &baseRepoStub{}
	svc := basesvc.NewBaseService(repo, lg)
	srv := NewServer(svc, &Config{SearchRadius: 1000})

	_, err := srv.CreateBase(context.Background(), &pb.CreateBaseRequest{
		Name:     "",
		Location: &pbCommon.Location{Lat: 55, Lon: 37},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestServer_FindNearest_NotFoundFallback(t *testing.T) {
	repo := &baseRepoStub{
		GetNearestFn: func(ctx context.Context, loc shared.Location, radius float64) (*base.BaseNearest, error) {
			return nil, basesvc.ErrBaseNotFound
		},
	}
	svc := basesvc.NewBaseService(repo, lg)
	srv := NewServer(svc, &Config{SearchRadius: 1000})

	resp, err := srv.FindNearest(context.Background(), &pb.FindNearestRequest{
		DroneLocation: &pbCommon.Location{Lat: 55, Lon: 37},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetFound() {
		t.Fatalf("expected found=false")
	}
}

func TestServer_GetBaseLocation_NotFound(t *testing.T) {
	repo := &baseRepoStub{
		GetByIDFn: func(ctx context.Context, id string) (*base.Base, error) {
			return nil, basesvc.ErrBaseNotFound
		},
	}
	svc := basesvc.NewBaseService(repo, lg)
	srv := NewServer(svc, &Config{SearchRadius: 1000})

	_, err := srv.GetBaseLocation(context.Background(), &pb.GetBaseLocationRequest{BaseId: "x"})
	if status.Code(err) != codes.NotFound && !errors.Is(err, basesvc.ErrBaseNotFound) {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
