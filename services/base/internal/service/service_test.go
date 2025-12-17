package service

import (
	"context"
	"errors"
	"hive/pkg/logger"
	"testing"

	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type baseRepoStub struct {
	SaveFn       func(context.Context, *base.Base) error
	GetByIDFn    func(context.Context, string) (*base.Base, error)
	ListFn       func(context.Context, int64, int64) ([]*base.Base, error)
	GetNearestFn func(context.Context, shared.Location, float64) (*base.BaseNearest, error)

	SaveCalls       int
	GetByIDCalls    int
	ListCalls       int
	GetNearestCalls int
	LastSaved       *base.Base
}

func (r *baseRepoStub) Save(ctx context.Context, b *base.Base) error {
	r.SaveCalls++
	r.LastSaved = b
	if r.SaveFn != nil {
		return r.SaveFn(ctx, b)
	}
	return nil
}

func (r *baseRepoStub) GetByID(ctx context.Context, id string) (*base.Base, error) {
	r.GetByIDCalls++
	if r.GetByIDFn != nil {
		return r.GetByIDFn(ctx, id)
	}
	return nil, ErrBaseNotFound
}

func (r *baseRepoStub) List(ctx context.Context, offset, limit int64) ([]*base.Base, error) {
	r.ListCalls++
	if r.ListFn != nil {
		return r.ListFn(ctx, offset, limit)
	}
	return []*base.Base{}, nil
}

func (r *baseRepoStub) GetNearest(ctx context.Context, location shared.Location, radius float64) (*base.BaseNearest, error) {
	r.GetNearestCalls++
	if r.GetNearestFn != nil {
		return r.GetNearestFn(ctx, location, radius)
	}
	return nil, ErrBaseNotFound
}

func TestBaseService_CreateBase_Validation(t *testing.T) {
	repo := &baseRepoStub{}
	svc := NewBaseService(repo, lg)

	_, err := svc.CreateBase(context.Background(), "", "addr", shared.Location{Lat: 55, Lon: 37})
	if err == nil {
		t.Fatalf("expected error")
	}
	if repo.SaveCalls != 0 {
		t.Fatalf("expected Save not called")
	}

	_, err = svc.CreateBase(context.Background(), "name", "addr", shared.Location{Lat: 200, Lon: 37})
	if err == nil {
		t.Fatalf("expected error")
	}
	if repo.SaveCalls != 0 {
		t.Fatalf("expected Save not called")
	}
}

func TestBaseService_CreateBase_SavesAndReturnsID(t *testing.T) {
	repo := &baseRepoStub{}
	svc := NewBaseService(repo, lg)

	id, err := svc.CreateBase(context.Background(), "base", "addr", shared.Location{Lat: 55.7, Lon: 37.6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected id")
	}
	if repo.SaveCalls != 1 {
		t.Fatalf("expected Save called once")
	}
	if repo.LastSaved == nil || repo.LastSaved.ID == "" || repo.LastSaved.Name != "base" {
		t.Fatalf("unexpected saved base")
	}
}

func TestBaseService_GetLocation_ErrorWrapped(t *testing.T) {
	repo := &baseRepoStub{
		GetByIDFn: func(ctx context.Context, id string) (*base.Base, error) {
			return nil, ErrBaseNotFound
		},
	}
	svc := NewBaseService(repo, lg)

	_, err := svc.GetLocation(context.Background(), "x")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrBaseNotFound) {
		t.Fatalf("expected ErrBaseNotFound, got %v", err)
	}
}

func TestBaseService_ListBases_LimitZero(t *testing.T) {
	repo := &baseRepoStub{}
	svc := NewBaseService(repo, lg)

	res, err := svc.ListBases(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected empty")
	}
	if repo.ListCalls != 0 {
		t.Fatalf("expected List not called")
	}
}

func TestBaseService_FindNearest_ErrorWrapped(t *testing.T) {
	repo := &baseRepoStub{
		GetNearestFn: func(ctx context.Context, loc shared.Location, radius float64) (*base.BaseNearest, error) {
			return nil, ErrBaseNotFound
		},
	}
	svc := NewBaseService(repo, lg)

	_, err := svc.FindNearest(context.Background(), shared.Location{Lat: 55, Lon: 37}, 1000)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrBaseNotFound) {
		t.Fatalf("expected ErrBaseNotFound")
	}
}
