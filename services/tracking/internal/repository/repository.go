package repository

import (
	"context"
	hredis "hive/pkg/db/redis"

	"github.com/redis/go-redis/v9"
)

type TrackingRepository struct {
	rdb *hredis.Database
}

func New(rdb *hredis.Database) *TrackingRepository {
	return &TrackingRepository{rdb: rdb}
}

func (t *TrackingRepository) UpdateGeo(ctx context.Context, droneID string, lon, lat float64) error {
	// TODO: сохранять геолокацию по ключу drones:geo
	return t.rdb.Client.GeoAdd(ctx, "drones",
		&redis.GeoLocation{
			Name:      droneID,
			Longitude: lon,
			Latitude:  lat,
		},
	).Err()
}

func (t *TrackingRepository) UpdateState(ctx context.Context, droneID string, battery float64, status string) error {
	// TODO: принимать структуру drone.TelemetryData и работать с ним
	key := "drone:" + droneID
	return t.rdb.Client.HSet(ctx, key, map[string]any{
		"battery": battery,
		"status":  status,
	}).Err()
}

func (t *TrackingRepository) GetDrone(ctx context.Context, droneID string) (map[string]string, error) {
	// TODO: маппить JSON в структуру drone.Drone
	return t.rdb.Client.HGetAll(ctx, "drone:"+droneID).Result()
}
