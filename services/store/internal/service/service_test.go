package service

import (
	"context"
	"errors"
	"hive/pkg/logger"
	"testing"

	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type storeRepoStub struct {
	SaveFn       func(context.Context, *store.Store) error
	GetByIDFn    func(context.Context, string) (*store.Store, error)
	ListFn       func(context.Context, int64, int64) ([]*store.Store, error)
	GetNearestFn func(context.Context, shared.Location, float64) (*store.StoreNearest, error)

	SaveCalls       int
	GetByIDCalls    int
	ListCalls       int
	GetNearestCalls int
	LastSaved       *store.Store
}

func (r *storeRepoStub) Save(ctx context.Context, s *store.Store) error {
	r.SaveCalls++
	r.LastSaved = s
	if r.SaveFn != nil {
		return r.SaveFn(ctx, s)
	}
	return nil
}

func (r *storeRepoStub) GetByID(ctx context.Context, id string) (*store.Store, error) {
	r.GetByIDCalls++
	if r.GetByIDFn != nil {
		return r.GetByIDFn(ctx, id)
	}
	return nil, ErrStoreNotFound
}

func (r *storeRepoStub) List(ctx context.Context, offset, limit int64) ([]*store.Store, error) {
	r.ListCalls++
	if r.ListFn != nil {
		return r.ListFn(ctx, offset, limit)
	}
	return []*store.Store{}, nil
}

func (r *storeRepoStub) GetNearest(ctx context.Context, location shared.Location, radius float64) (*store.StoreNearest, error) {
	r.GetNearestCalls++
	if r.GetNearestFn != nil {
		return r.GetNearestFn(ctx, location, radius)
	}
	return nil, ErrStoreNotFound
}

func TestStoreService_CreateStore_Validation(t *testing.T) {
	repo := &storeRepoStub{}
	svc := NewStoreService(repo, lg)

	_, err := svc.CreateStore(context.Background(), "", "addr", shared.Location{Lat: 55, Lon: 37})
	if err == nil {
		t.Fatalf("expected error")
	}
	if repo.SaveCalls != 0 {
		t.Fatalf("expected Save not called")
	}

	_, err = svc.CreateStore(context.Background(), "name", "addr", shared.Location{Lat: -200, Lon: 37})
	if err == nil {
		t.Fatalf("expected error")
	}
	if repo.SaveCalls != 0 {
		t.Fatalf("expected Save not called")
	}
}

func TestStoreService_CreateStore_SavesAndReturnsID(t *testing.T) {
	repo := &storeRepoStub{}
	svc := NewStoreService(repo, lg)

	id, err := svc.CreateStore(context.Background(), "store", "addr", shared.Location{Lat: 55.7, Lon: 37.6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected id")
	}
	if repo.SaveCalls != 1 {
		t.Fatalf("expected Save called once")
	}
	if repo.LastSaved == nil || repo.LastSaved.ID == "" || repo.LastSaved.Name != "store" {
		t.Fatalf("unexpected saved store")
	}
}

func TestStoreService_GetLocation_ErrorWrapped(t *testing.T) {
	repo := &storeRepoStub{
		GetByIDFn: func(ctx context.Context, id string) (*store.Store, error) {
			return nil, ErrStoreNotFound
		},
	}
	svc := NewStoreService(repo, lg)

	_, err := svc.GetLocation(context.Background(), "x")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("expected ErrStoreNotFound")
	}
}

func TestStoreService_ListStores_LimitZero(t *testing.T) {
	repo := &storeRepoStub{}
	svc := NewStoreService(repo, lg)

	res, err := svc.ListStores(context.Background(), 0, 0)
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
