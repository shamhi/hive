package repository

import (
	"context"
	"encoding/json"
	"fmt"
	hredis "hive/pkg/db/redis"
	"hive/services/tracking/internal/models"

	"github.com/redis/go-redis/v9"
)

const (
	AllDronesKey  string = "drones:all"
	DroneDataKey  string = "drone:data"
	DroneGeoKey   string = "drones:geo"
	GeoSearchUnit string = "m"
)

type TrackingRepository struct {
	rdb *hredis.Database
}

func New(rdb *hredis.Database) *TrackingRepository {
	return &TrackingRepository{rdb: rdb}
}

func (t *TrackingRepository) FindNearest(ctx context.Context, location models.Location, radius float64) ([]redis.GeoLocation, error) {
	res, err := t.rdb.Client.GeoRadius(ctx, DroneGeoKey, location.Lon, location.Lat, &redis.GeoRadiusQuery{
		Radius: radius,
		Unit:   GeoSearchUnit,
		Sort:   "ASC",
		Count:  1,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to find nearest drone: %w", err)
	}

	return res, nil
}

func (t *TrackingRepository) GetData(ctx context.Context, droneID string) (map[string]string, error) {
	return t.rdb.Client.HGetAll(ctx, DroneDataKey+":"+droneID).Result()
}

func (t *TrackingRepository) GetGeopostion(ctx context.Context, droneId string) (models.Location, error) {
	positions, err := t.rdb.Client.GeoPos(ctx, DroneGeoKey, droneId).Result()
	if err != nil {
		return models.Location{}, fmt.Errorf("failed to get geoposition: %w", err)
	}

	if len(positions) == 0 || positions[0] == nil {
		return models.Location{}, fmt.Errorf("geoposition not found for drone %s", droneId)
	}

	pos := positions[0]

	return models.Location{
		Lat: pos.Latitude,
		Lon: pos.Longitude,
	}, nil
}

func (t *TrackingRepository) SetStatus(ctx context.Context, droneID string, status int32) error {
	key := DroneDataKey + ":" + droneID
	data, err := t.rdb.Client.HGet(ctx, key, "data").Result()
	if err != nil {
		return fmt.Errorf("failed to get drone data: %w", err)
	}

	var drone models.Drone
	if err := json.Unmarshal([]byte(data), &drone); err != nil {
		return fmt.Errorf("failed to deserialize drone data: %w", err)
	}

	drone.Status = models.DroneStatus(status)

	updatedData, err := json.Marshal(drone)
	if err != nil {
		return fmt.Errorf("failed to serialize drone data: %w", err)
	}

	return t.rdb.Client.HSet(ctx, key, "data", updatedData).Err()
}

func (t *TrackingRepository) UpdateAllDrones(ctx context.Context, drone models.Drone) error {
	return t.rdb.Client.SAdd(ctx, AllDronesKey, drone.ID).Err()
}

func (t *TrackingRepository) UpdateData(ctx context.Context, drone models.Drone) error {
	key := DroneDataKey + ":" + drone.ID
	data, err := json.Marshal(drone)

	if err != nil {
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	return t.rdb.Client.HSet(ctx, key, data).Err()
}

func (t *TrackingRepository) UpdateGeopostion(ctx context.Context, drone models.Drone) error {
	return t.rdb.Client.GeoAdd(ctx, DroneGeoKey,
		&redis.GeoLocation{
			Name:      drone.ID,
			Longitude: drone.Location.Lon,
			Latitude:  drone.Location.Lat,
		},
	).Err()
}
