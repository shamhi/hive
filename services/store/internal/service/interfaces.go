package service

import (
	"context"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"
)

type StoreRepository interface {
	Save(ctx context.Context, s *store.Store) error
	GetByID(ctx context.Context, id string) (*store.Store, error)
	GetNearest(ctx context.Context, deliveryLocation shared.Location, radiusMeters float64) (*store.StoreNearest, error)
}
