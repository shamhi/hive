package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"hive/pkg/logger"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/service"
	"time"

	"go.uber.org/zap"
)

type Handler struct {
	repo service.DroneRepository
	lg   logger.Logger
}

func New(
	repo service.DroneRepository,
	lg logger.Logger,
) *Handler {
	return &Handler{
		repo: repo,
		lg:   lg,
	}
}

func (h *Handler) HandleMessage(ctx context.Context, data []byte) error {
	start := time.Now()

	lg := h.lg.With(
		zap.String("component", "tracking_kafka_handler"),
		zap.String("op", "HandleMessage"),
		zap.Int("payload_bytes", len(data)),
	)

	var tm drone.TelemetryData
	if err := json.Unmarshal(data, &tm); err != nil {
		lg.Error(ctx, "failed to unmarshal telemetry data",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		return err
	}

	lg = lg.With(
		zap.String("drone_id", tm.DroneID),
		zap.Int64("timestamp", tm.Timestamp),
	)

	lg.Info(ctx, "telemetry message received",
		zap.Float64("lat", tm.DroneLocation.Lat),
		zap.Float64("lon", tm.DroneLocation.Lon),
		zap.Float64("battery", tm.Battery),
		zap.Float64("speed_mps", tm.SpeedMps),
		zap.Float64("consumption_per_meter", tm.ConsumptionPerMeter),
		zap.String("status", string(tm.Status)),
	)

	if tm.DroneID == "" {
		err := fmt.Errorf("drone_id is empty")
		lg.Warn(ctx, "invalid telemetry message",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return err
	}

	if err := h.repo.UpdateState(ctx, tm); err != nil {
		lg.Error(ctx, "failed to update drone state",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		return err
	}

	lg.Info(ctx, "drone state updated", zap.Duration("duration", time.Since(start)))
	return nil
}
