package service

import (
	"context"
	"fmt"
	"hive/pkg/logger"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BaseService struct {
	repo BaseRepository
	lg   logger.Logger
}

func NewBaseService(
	repo BaseRepository,
	lg logger.Logger,
) *BaseService {
	return &BaseService{
		repo: repo,
		lg:   lg,
	}
}

func (s *BaseService) CreateBase(
	ctx context.Context,
	name string,
	address string,
	location shared.Location,
) (string, error) {
	lg := s.lg.With(
		zap.String("component", "base_service"),
		zap.String("op", "CreateBase"),
		zap.String("name", name),
	)

	start := time.Now()

	if name == "" {
		lg.Warn(ctx, "validation failed: base name is empty")
		return "", fmt.Errorf("base name cannot be empty")
	}
	if location.Lat < -90 || location.Lat > 90 || location.Lon < -180 || location.Lon > 180 {
		lg.Warn(ctx, "validation failed: invalid coordinates",
			zap.Float64("lat", location.Lat),
			zap.Float64("lon", location.Lon),
		)
		return "", fmt.Errorf("invalid base location coordinates")
	}

	lg.Info(ctx, "create base started",
		zap.String("address", address),
		zap.Float64("lat", location.Lat),
		zap.Float64("lon", location.Lon),
	)

	b := &base.Base{
		ID:       uuid.NewString(),
		Name:     name,
		Address:  address,
		Location: location,
	}

	if err := s.repo.Save(ctx, b); err != nil {
		lg.Error(ctx, "failed to save base", zap.String("base_id", b.ID), zap.Error(err), zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to save base: %w", err)
	}

	lg.Info(ctx, "create base completed",
		zap.String("base_id", b.ID),
		zap.Duration("duration", time.Since(start)),
	)

	return b.ID, nil
}

func (s *BaseService) GetLocation(
	ctx context.Context,
	id string,
) (*base.Base, error) {
	lg := s.lg.With(
		zap.String("component", "base_service"),
		zap.String("op", "GetLocation"),
		zap.String("base_id", id),
	)

	start := time.Now()
	if id == "" {
		lg.Warn(ctx, "validation failed: base_id is empty")
		return nil, fmt.Errorf("base id cannot be empty")
	}

	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		lg.Error(ctx, "failed to get base by ID", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get base by ID: %w", err)
	}

	lg.Info(ctx, "get location completed",
		zap.String("name", b.Name),
		zap.String("address", b.Address),
		zap.Float64("lat", b.Location.Lat),
		zap.Float64("lon", b.Location.Lon),
		zap.Duration("duration", time.Since(start)),
	)

	return b, nil
}

func (s *BaseService) ListBases(
	ctx context.Context,
	offset, limit int64,
) ([]*base.Base, error) {
	lg := s.lg.With(
		zap.String("component", "base_service"),
		zap.String("op", "ListBases"),
		zap.Int64("offset", offset),
		zap.Int64("limit", limit),
	)

	start := time.Now()
	if limit <= 0 {
		lg.Info(ctx, "limit <= 0, returning empty list", zap.Duration("duration", time.Since(start)))
		return []*base.Base{}, nil
	}
	if offset < 0 {
		lg.Warn(ctx, "offset < 0, treating as 0", zap.Int64("offset", offset))
		offset = 0
	}

	bases, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		lg.Error(ctx, "failed to list bases", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to list bases: %w", err)
	}

	lg.Info(ctx, "list bases completed",
		zap.Int("count", len(bases)),
		zap.Duration("duration", time.Since(start)),
	)

	return bases, nil
}

func (s *BaseService) FindNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
) (*base.BaseNearest, error) {
	lg := s.lg.With(
		zap.String("component", "base_service"),
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

	bNearest, err := s.repo.GetNearest(ctx, location, radiusMeters)
	if err != nil {
		lg.Error(ctx, "failed to find nearest base", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to find nearest base: %w", err)
	}

	lg.Info(ctx, "find nearest completed",
		zap.String("base_id", bNearest.ID),
		zap.Float64("distance", bNearest.Distance),
		zap.Duration("duration", time.Since(start)),
	)

	return bNearest, nil
}
