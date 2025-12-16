package redis

import (
	"context"
	"fmt"
	"hive/services/base/internal/domain/base"
	"hive/services/base/internal/domain/shared"
	"hive/services/base/internal/service"

	"github.com/redis/go-redis/v9"
)

const (
	AllBasesKey   string = "bases:all"
	BaseDataKey   string = "bases:data:"
	BaseGeoKey    string = "bases:geo"
	GeoSearchUnit string = "m"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) Save(ctx context.Context, s *base.Base) error {
	pipe := r.client.TxPipeline()

	pipe.SAdd(ctx, AllBasesKey, s.ID)

	pipe.HSet(ctx, BaseDataKey+s.ID,
		"name", s.Name,
		"address", s.Address,
	)

	pipe.GeoAdd(ctx, BaseGeoKey,
		&redis.GeoLocation{
			Name:      s.ID,
			Longitude: s.Location.Lon,
			Latitude:  s.Location.Lat,
		},
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to save base to redis: %w", err)
	}

	return nil
}

func (r *RedisRepo) GetByID(ctx context.Context, baseID string) (*base.Base, error) {
	data, err := r.client.HGetAll(ctx, BaseDataKey+baseID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get base data: %w", err)
	}

	if len(data) == 0 {
		return nil, service.ErrBaseNotFound
	}

	name := data["name"]
	address := data["address"]

	positions, err := r.client.GeoPos(ctx, BaseGeoKey, baseID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get base geoposition: %w", err)
	}

	if len(positions) == 0 || positions[0] == nil {
		return nil, service.ErrBaseNotFound
	}

	pos := positions[0]

	return &base.Base{
		ID:      baseID,
		Name:    name,
		Address: address,
		Location: shared.Location{
			Lat: pos.Latitude,
			Lon: pos.Longitude,
		},
	}, nil
}

func (r *RedisRepo) List(ctx context.Context, offset, limit int64) ([]*base.Base, error) {
	start := offset
	stop := offset + limit - 1

	ids, err := r.client.ZRange(ctx, AllBasesKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get base ids: %w", err)
	}

	if len(ids) == 0 {
		return []*base.Base{}, nil
	}

	pipe := r.client.Pipeline()
	hCmd := make([]*redis.MapStringStringCmd, 0, len(ids))
	gCmd := make([]*redis.GeoPosCmd, 0, len(ids))

	for _, id := range ids {
		hCmd = append(hCmd, pipe.HGetAll(ctx, BaseDataKey+id))
		gCmd = append(gCmd, pipe.GeoPos(ctx, BaseGeoKey, id))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to fetch bases data: %w", err)
	}

	bases := make([]*base.Base, 0, len(ids))
	for i, id := range ids {
		data, err := hCmd[i].Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get base data: %w", err)
		}

		if len(data) == 0 {
			continue
		}

		name := data["name"]
		address := data["address"]

		posArr, err := gCmd[i].Result()
		if err != nil || len(posArr) == 0 || posArr[0] == nil {
			continue
		}

		bases = append(bases, &base.Base{
			ID:      id,
			Name:    name,
			Address: address,
			Location: shared.Location{
				Lat: posArr[0].Latitude,
				Lon: posArr[0].Longitude,
			},
		})
	}

	return bases, nil
}

func (r *RedisRepo) GetNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
) (*base.BaseNearest, error) {
	result, err := r.client.GeoSearchLocation(
		ctx,
		BaseGeoKey,
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
		return nil, service.ErrBaseNotFound
	}

	nearestResult := result[0]
	return &base.BaseNearest{
		ID:       nearestResult.Name,
		Distance: nearestResult.Dist,
	}, nil
}
