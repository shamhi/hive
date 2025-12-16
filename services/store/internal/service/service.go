package service

import (
	"context"
	"fmt"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"

	"github.com/google/uuid"
)

type StoreService struct {
	repo StoreRepository
}

func NewStoreService(repo StoreRepository) *StoreService {
	return &StoreService{repo: repo}
}

func (s *StoreService) CreateStore(
	ctx context.Context,
	name string,
	address string,
	location shared.Location,
) (string, error) {
	if name == "" {
		return "", fmt.Errorf("store name cannot be empty")
	}
	if location.Lat < -90 || location.Lat > 90 || location.Lon < -180 || location.Lon > 180 {
		return "", fmt.Errorf("invalid store location coordinates")
	}

	st := &store.Store{
		ID:       uuid.NewString(),
		Name:     name,
		Address:  address,
		Location: location,
	}
	if err := s.repo.Save(ctx, st); err != nil {
		return "", fmt.Errorf("failed to save store: %w", err)
	}

	return st.ID, nil
}

func (s *StoreService) GetLocation(
	ctx context.Context,
	id string,
) (*store.Store, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get store by ID: %w", err)
	}

	return st, nil
}

func (s *StoreService) ListStores(
	ctx context.Context,
	offset, limit int64,
) ([]*store.Store, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	stores, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}

	return stores, nil
}

func (s *StoreService) FindNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
) (*store.StoreNearest, error) {
	stNearest, err := s.repo.GetNearest(ctx, location, radiusMeters)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest store: %w", err)
	}

	return stNearest, nil
}
