package redis

import (
	"context"
	"fmt"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"
	"hive/services/store/internal/service"

	"github.com/redis/go-redis/v9"
)

const (
	AllStoresKey  string = "stores:all"
	StoreDataKey  string = "stores:data:"
	StoreGeoKey   string = "stores:geo"
	GeoSearchUnit string = "m"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) Save(ctx context.Context, s *store.Store) error {
	pipe := r.client.TxPipeline()

	pipe.SAdd(ctx, AllStoresKey, s.ID)

	pipe.HSet(ctx, StoreDataKey+s.ID,
		"name", s.Name,
		"address", s.Address,
	)

	pipe.GeoAdd(ctx, StoreGeoKey,
		&redis.GeoLocation{
			Name:      s.ID,
			Longitude: s.Location.Lon,
			Latitude:  s.Location.Lat,
		},
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to save store to redis: %w", err)
	}

	return nil
}

func (r *RedisRepo) GetByID(ctx context.Context, storeID string) (*store.Store, error) {
	data, err := r.client.HGetAll(ctx, StoreDataKey+storeID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get store data: %w", err)
	}

	if len(data) == 0 {
		return nil, service.ErrStoreNotFound
	}

	name := data["name"]
	address := data["address"]

	positions, err := r.client.GeoPos(ctx, StoreGeoKey, storeID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get store geoposition: %w", err)
	}

	if len(positions) == 0 || positions[0] == nil {
		return nil, service.ErrStoreNotFound
	}

	pos := positions[0]

	return &store.Store{
		ID:      storeID,
		Name:    name,
		Address: address,
		Location: shared.Location{
			Lat: pos.Latitude,
			Lon: pos.Longitude,
		},
	}, nil
}

func (r *RedisRepo) GetNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
) (*store.StoreNearest, error) {
	result, err := r.client.GeoSearchLocation(
		ctx,
		StoreGeoKey,
		&redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  location.Lon,
				Latitude:   location.Lat,
				Radius:     radiusMeters,
				RadiusUnit: GeoSearchUnit,
				Sort:       "ASC",
				Count:      1,
			},
			WithDist: true,
		},
	).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, service.ErrStoreNotFound
	}

	nearestResult := result[0]
	return &store.StoreNearest{
		ID:       nearestResult.Name,
		Distance: nearestResult.Dist,
	}, nil
}
