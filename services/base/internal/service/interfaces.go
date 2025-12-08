package service

import (
	"context"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"
)

type BaseRepository interface {
	Save(ctx context.Context, s *base.Base) error
	GetByID(ctx context.Context, id string) (*base.Base, error)
	GetNearest(ctx context.Context, location shared.Location, radiusMeters float64) (*base.BaseNearest, error)
}
