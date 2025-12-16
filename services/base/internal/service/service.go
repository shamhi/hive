package service

import (
	"context"
	"fmt"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"

	"github.com/google/uuid"
)

type BaseService struct {
	repo BaseRepository
}

func NewBaseService(repo BaseRepository) *BaseService {
	return &BaseService{repo: repo}
}

func (s *BaseService) CreateBase(
	ctx context.Context,
	name string,
	address string,
	location shared.Location,
) (string, error) {
	if name == "" {
		return "", fmt.Errorf("base name cannot be empty")
	}
	if location.Lat < -90 || location.Lat > 90 || location.Lon < -180 || location.Lon > 180 {
		return "", fmt.Errorf("invalid base location coordinates")
	}

	st := &base.Base{
		ID:       uuid.NewString(),
		Name:     name,
		Address:  address,
		Location: location,
	}
	if err := s.repo.Save(ctx, st); err != nil {
		return "", fmt.Errorf("failed to save base: %w", err)
	}

	return st.ID, nil
}

func (s *BaseService) GetLocation(
	ctx context.Context,
	id string,
) (*base.Base, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get base by ID: %w", err)
	}

	return st, nil
}

func (s *BaseService) ListBases(
	ctx context.Context,
	offset, limit int64,
) ([]*base.Base, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	bases, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list bases: %w", err)
	}

	return bases, nil
}

func (s *BaseService) FindNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
) (*base.BaseNearest, error) {
	stNearest, err := s.repo.GetNearest(ctx, location, radiusMeters)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest base: %w", err)
	}

	return stNearest, nil
}
