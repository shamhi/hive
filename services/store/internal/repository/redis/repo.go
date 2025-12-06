package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"hive/services/store/internal/domain/shared"
	"hive/services/store/internal/domain/store"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) Save(ctx context.Context, s *store.Store) error {
	if err := r.client.GeoAdd(
		ctx,
		"stores:geo",
		&redis.GeoLocation{
			Name:      s.ID,
			Longitude: s.Location.Lon,
			Latitude:  s.Location.Lat,
		},
	).Err(); err != nil {
		return err
	}

	storeJSON, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if err := r.client.Set(
		ctx,
		fmt.Sprintf("stores:data:%s", s.ID),
		storeJSON,
		0,
	).Err(); err != nil {
		return err
	}

	if err := r.client.SAdd(
		ctx,
		"stores:all",
		s.ID,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisRepo) GetByID(ctx context.Context, id string) (*store.Store, error) {
	storeJSON, err := r.client.Get(ctx, fmt.Sprintf("stores:data:%s", id)).Result()
	if err != nil {
		return nil, err
	}

	var s store.Store
	if err := json.Unmarshal([]byte(storeJSON), &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *RedisRepo) GetNearest(
	ctx context.Context,
	deliveryLocation shared.Location,
	radiusMeters float64,
) (*store.StoreNearest, error) {
	result, err := r.client.GeoSearchLocation(
		ctx,
		"stores:geo",
		&redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  deliveryLocation.Lon,
				Latitude:   deliveryLocation.Lat,
				Radius:     radiusMeters,
				RadiusUnit: "m",
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
		return nil, fmt.Errorf("no stores found within radius")
	}

	nearestResult := result[0]
	return &store.StoreNearest{
		ID:       nearestResult.Name,
		Distance: nearestResult.Dist,
	}, nil
}
