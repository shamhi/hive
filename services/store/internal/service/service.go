package service

import (
	"context"
	"fmt"
	"hive/pkg/logger"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type StoreService struct {
	repo StoreRepository
	lg   logger.Logger
}

func NewStoreService(
	repo StoreRepository,
	lg logger.Logger,
) *StoreService {
	return &StoreService{
		repo: repo,
		lg:   lg,
	}
}

func (s *StoreService) CreateStore(
	ctx context.Context,
	name string,
	address string,
	location shared.Location,
) (string, error) {
	lg := s.lg.With(
		zap.String("component", "store_service"),
		zap.String("op", "CreateStore"),
		zap.String("name", name),
	)

	start := time.Now()

	if name == "" {
		lg.Warn(ctx, "validation failed: store name is empty")
		return "", fmt.Errorf("store name cannot be empty")
	}
	if location.Lat < -90 || location.Lat > 90 || location.Lon < -180 || location.Lon > 180 {
		lg.Warn(ctx, "validation failed: invalid coordinates",
			zap.Float64("lat", location.Lat),
			zap.Float64("lon", location.Lon),
		)
		return "", fmt.Errorf("invalid store location coordinates")
	}

	lg.Info(ctx, "create store started",
		zap.String("address", address),
		zap.Float64("lat", location.Lat),
		zap.Float64("lon", location.Lon),
	)

	st := &store.Store{
		ID:       uuid.NewString(),
		Name:     name,
		Address:  address,
		Location: location,
	}

	if err := s.repo.Save(ctx, st); err != nil {
		lg.Error(ctx, "failed to save store", zap.String("store_id", st.ID), zap.Error(err), zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to save store: %w", err)
	}

	lg.Info(ctx, "create store completed",
		zap.String("store_id", st.ID),
		zap.Duration("duration", time.Since(start)),
	)

	return st.ID, nil
}

func (s *StoreService) GetLocation(
	ctx context.Context,
	id string,
) (*store.Store, error) {
	lg := s.lg.With(
		zap.String("component", "store_service"),
		zap.String("op", "GetLocation"),
		zap.String("store_id", id),
	)

	start := time.Now()
	if id == "" {
		lg.Warn(ctx, "validation failed: store_id is empty")
		return nil, fmt.Errorf("store id cannot be empty")
	}

	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		lg.Error(ctx, "failed to get store by ID", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get store by ID: %w", err)
	}

	lg.Info(ctx, "get location completed",
		zap.String("name", st.Name),
		zap.String("address", st.Address),
		zap.Float64("lat", st.Location.Lat),
		zap.Float64("lon", st.Location.Lon),
		zap.Duration("duration", time.Since(start)),
	)

	return st, nil
}

func (s *StoreService) ListStores(
	ctx context.Context,
	offset, limit int64,
) ([]*store.Store, error) {
	lg := s.lg.With(
		zap.String("component", "store_service"),
		zap.String("op", "ListStores"),
		zap.Int64("offset", offset),
		zap.Int64("limit", limit),
	)

	start := time.Now()
	if limit <= 0 {
		lg.Info(ctx, "limit <= 0, returning empty list", zap.Duration("duration", time.Since(start)))
		return []*store.Store{}, nil
	}
	if offset < 0 {
		lg.Warn(ctx, "offset < 0, treating as 0", zap.Int64("offset", offset))
		offset = 0
	}

	stores, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		lg.Error(ctx, "failed to list stores", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}

	lg.Info(ctx, "list stores completed",
		zap.Int("count", len(stores)),
		zap.Duration("duration", time.Since(start)),
	)

	return stores, nil
}

func (s *StoreService) FindNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
) (*store.StoreNearest, error) {
	lg := s.lg.With(
		zap.String("component", "store_service"),
		zap.String("op", "FindNearest"),
		zap.Float64("lat", location.Lat),
		zap.Float64("lon", location.Lon),
		zap.Float64("radius_meters", radiusMeters),
	)

	start := time.Now()
	if radiusMeters <= 0 {
		lg.Warn(ctx, "validation failed: radiusMeters <= 0", zap.Float64("radius_meters", radiusMeters))
		return nil, fmt.Errorf("radius must be positive")
	}

	stNearest, err := s.repo.GetNearest(ctx, location, radiusMeters)
	if err != nil {
		lg.Error(ctx, "failed to find nearest store", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to find nearest store: %w", err)
	}

	lg.Info(ctx, "find nearest completed",
		zap.String("store_id", stNearest.ID),
		zap.Float64("distance", stNearest.Distance),
		zap.Duration("duration", time.Since(start)),
	)

	return stNearest, nil
}
