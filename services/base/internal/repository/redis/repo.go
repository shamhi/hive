package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) Save(ctx context.Context, s *base.Base) error {
	if err := r.client.GeoAdd(
		ctx,
		"bases:geo",
		&redis.GeoLocation{
			Name:      s.ID,
			Longitude: s.Location.Lon,
			Latitude:  s.Location.Lat,
		},
	).Err(); err != nil {
		return err
	}

	baseJSON, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if err := r.client.Set(
		ctx,
		fmt.Sprintf("bases:data:%s", s.ID),
		baseJSON,
		0,
	).Err(); err != nil {
		return err
	}

	if err := r.client.SAdd(
		ctx,
		"bases:all",
		s.ID,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisRepo) GetByID(ctx context.Context, id string) (*base.Base, error) {
	baseJSON, err := r.client.Get(ctx, fmt.Sprintf("bases:data:%s", id)).Result()
	if err != nil {
		return nil, err
	}

	var s base.Base
	if err := json.Unmarshal([]byte(baseJSON), &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *RedisRepo) GetNearest(
	ctx context.Context,
	droneLocation shared.Location,
	radiusMeters float64,
) (*base.BaseNearest, error) {
	result, err := r.client.GeoSearchLocation(
		ctx,
		"bases:geo",
		&redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  droneLocation.Lon,
				Latitude:   droneLocation.Lat,
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
		return nil, fmt.Errorf("no bases found within radius")
	}

	nearestResult := result[0]
	return &base.BaseNearest{
		ID:       nearestResult.Name,
		Distance: nearestResult.Dist,
	}, nil
}
