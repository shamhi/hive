package redis

import (
	"context"
	"fmt"
	hredis "hive/pkg/db/redis"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/domain/shared"
	"hive/services/tracking/internal/service"
	"slices"
	"sort"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	AllDronesKey         string = "drones:all"
	DroneDataKey         string = "drones:data:"
	DroneGeoKey          string = "drones:geo"
	GeoSearchUnit        string = "m"
	MaxSearchDronesCount int    = 1000
)

type RedisRepo struct {
	rdb *hredis.Database
}

func NewRedisRepo(rdb *hredis.Database) *RedisRepo {
	return &RedisRepo{rdb: rdb}
}

func (r *RedisRepo) GetNearest(
	ctx context.Context,
	location shared.Location,
	radiusMeters float64,
	minBattery float64,
) (*drone.DroneNearest, error) {
	result, err := r.rdb.Client.GeoSearchLocation(
		ctx,
		DroneGeoKey,
		&redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  location.Lon,
				Latitude:   location.Lat,
				Radius:     radiusMeters,
				RadiusUnit: GeoSearchUnit,
				Sort:       "ASC",
				Count:      MaxSearchDronesCount,
			},
			WithDist: true,
		},
	).Result()
	if err != nil {
		return nil, err
	}

	for _, res := range result {
		data, err := r.rdb.Client.HGetAll(ctx, DroneDataKey+res.Name).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get drone data: %w", err)
		}

		battery, _ := strconv.ParseFloat(data["battery"], 64)
		status := drone.DroneStatus(data["status"])

		if battery < minBattery || !slices.Contains([]drone.DroneStatus{
			drone.DroneStatusFree,
			drone.DroneStatusCharging,
		}, status) {
			continue
		}

		return &drone.DroneNearest{
			ID:       res.Name,
			Distance: res.Dist,
		}, nil
	}

	return nil, service.ErrDroneNotFound
}

func (r *RedisRepo) GetByID(
	ctx context.Context,
	droneID string,
) (*drone.Drone, error) {
	data, err := r.rdb.Client.HGetAll(ctx, DroneDataKey+droneID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get drone data: %w", err)
	}

	if len(data) == 0 {
		return nil, service.ErrDroneNotFound
	}

	battery, _ := strconv.ParseFloat(data["battery"], 64)
	speed, _ := strconv.ParseFloat(data["speed_mps"], 64)
	consumption, _ := strconv.ParseFloat(data["consumption_per_meter"], 64)
	status := drone.DroneStatus(data["status"])
	ts, _ := strconv.ParseInt(data["timestamp"], 10, 64)

	positions, err := r.rdb.Client.GeoPos(ctx, DroneGeoKey, droneID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get drone geoposition: %w", err)
	}

	if len(positions) == 0 || positions[0] == nil {
		return nil, service.ErrDroneNotFound
	}

	pos := positions[0]

	return &drone.Drone{
		ID:                  droneID,
		Battery:             battery,
		SpeedMps:            speed,
		ConsumptionPerMeter: consumption,
		Status:              status,
		Location: shared.Location{
			Lat: pos.Latitude,
			Lon: pos.Longitude,
		},
		UpdatedAt: ts,
	}, nil
}

func (r *RedisRepo) List(
	ctx context.Context,
	offset, limit int64,
) ([]*drone.Drone, error) {
	if limit <= 0 {
		return []*drone.Drone{}, nil
	}
	if offset < 0 {
		offset = 0
	}

	ids, err := r.rdb.Client.SMembers(ctx, AllDronesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get drone ids: %w", err)
	}
	if len(ids) == 0 {
		return []*drone.Drone{}, nil
	}

	sort.Strings(ids)

	if offset >= int64(len(ids)) {
		return []*drone.Drone{}, nil
	}

	end := offset + limit
	if end > int64(len(ids)) {
		end = int64(len(ids))
	}

	pageIDs := ids[offset:end]

	pipe := r.rdb.Client.Pipeline()
	hCmd := make([]*redis.MapStringStringCmd, 0, len(pageIDs))
	gCmd := make([]*redis.GeoPosCmd, 0, len(pageIDs))

	for _, id := range pageIDs {
		hCmd = append(hCmd, pipe.HGetAll(ctx, DroneDataKey+id))
		gCmd = append(gCmd, pipe.GeoPos(ctx, DroneGeoKey, id))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to fetch drones data: %w", err)
	}

	drones := make([]*drone.Drone, 0, len(pageIDs))
	for i, id := range pageIDs {
		data, err := hCmd[i].Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get drone data: %w", err)
		}
		if len(data) == 0 {
			continue
		}

		battery, _ := strconv.ParseFloat(data["battery"], 64)
		speed, _ := strconv.ParseFloat(data["speed_mps"], 64)
		consumption, _ := strconv.ParseFloat(data["consumption_per_meter"], 64)
		status := drone.DroneStatus(data["status"])
		ts, _ := strconv.ParseInt(data["timestamp"], 10, 64)

		posArr, err := gCmd[i].Result()
		if err != nil || len(posArr) == 0 || posArr[0] == nil {
			continue
		}

		drones = append(drones, &drone.Drone{
			ID:                  id,
			Battery:             battery,
			SpeedMps:            speed,
			ConsumptionPerMeter: consumption,
			Status:              status,
			Location: shared.Location{
				Lat: posArr[0].Latitude,
				Lon: posArr[0].Longitude,
			},
			UpdatedAt: ts,
		})
	}

	return drones, nil
}

func (r *RedisRepo) SetStatus(
	ctx context.Context,
	droneID string,
	status drone.DroneStatus,
) error {
	exists, err := r.rdb.Client.Exists(ctx, DroneDataKey+droneID).Result()
	if err != nil {
		return fmt.Errorf("failed to check drone existence: %w", err)
	}
	if exists == 0 {
		return service.ErrDroneNotFound
	}

	if err := r.rdb.Client.HSet(ctx, DroneDataKey+droneID,
		"status",
		string(status),
	).Err(); err != nil {
		return fmt.Errorf("failed to set drone status: %w", err)
	}

	return nil
}

func (r *RedisRepo) UpdateState(
	ctx context.Context,
	tm drone.TelemetryData,
) error {
	pipe := r.rdb.Client.TxPipeline()

	pipe.SAdd(ctx, AllDronesKey, tm.DroneID)

	pipe.HSet(ctx, DroneDataKey+tm.DroneID,
		"battery", tm.Battery,
		"speed_mps", tm.SpeedMps,
		"consumption_per_meter", tm.ConsumptionPerMeter,
		"status", string(tm.Status),
		"timestamp", tm.Timestamp,
	)

	pipe.GeoAdd(ctx, DroneGeoKey,
		&redis.GeoLocation{
			Name:      tm.DroneID,
			Longitude: tm.DroneLocation.Lon,
			Latitude:  tm.DroneLocation.Lat,
		},
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to update drone state: %w", err)
	}

	return nil
}
