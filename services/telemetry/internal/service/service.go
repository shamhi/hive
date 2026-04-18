package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"hive/services/telemetry/internal/domain/drone"
	"net"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type DroneConnection struct {
	Commands chan *drone.ServerCommand
}

type TelemetryService struct {
	eventsWriter *kafka.Writer
	dataWriter   *kafka.Writer
	eventsTopic  string
	dataTopic    string

	mu    sync.RWMutex
	conns map[string]*DroneConnection

	lg logger.Logger
}

var telemetryRetryCfg = resilience.RetryConfig{
	MaxAttempts: 4,
	BaseDelay:   80 * time.Millisecond,
	MaxDelay:    800 * time.Millisecond,
	Jitter:      0.2,
}

func NewTelemetryService(
	eventsWriter, dataWriter *kafka.Writer,
	eventsTopic, dataTopic string,
	lg logger.Logger,
) *TelemetryService {
	return &TelemetryService{
		eventsWriter: eventsWriter,
		dataWriter:   dataWriter,
		eventsTopic:  eventsTopic,
		dataTopic:    dataTopic,
		conns:        make(map[string]*DroneConnection),
		lg:           lg,
	}
}

func (s *TelemetryService) RegisterConnection(droneID string) *DroneConnection {
	lg := s.lg.With(
		zap.String("component", "telemetry_service"),
		zap.String("op", "RegisterConnection"),
		zap.String("drone_id", droneID),
	)

	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if old, ok := s.conns[droneID]; ok {
		close(old.Commands)
		delete(s.conns, droneID)
		lg.Warn(context.Background(), "replaced existing connection")
	}

	conn := &DroneConnection{Commands: make(chan *drone.ServerCommand, 16)}
	s.conns[droneID] = conn

	lg.Info(context.Background(), "connection registered",
		zap.Int("commands_buf", 16),
		zap.Duration("duration", time.Since(start)),
	)
	return conn
}

func (s *TelemetryService) UnregisterConnection(droneID string) {
	lg := s.lg.With(
		zap.String("component", "telemetry_service"),
		zap.String("op", "UnregisterConnection"),
		zap.String("drone_id", droneID),
	)

	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.conns[droneID]; exists {
		close(conn.Commands)
		delete(s.conns, droneID)
		lg.Info(context.Background(), "connection unregistered", zap.Duration("duration", time.Since(start)))
		return
	}

	lg.Info(context.Background(), "connection not found", zap.Duration("duration", time.Since(start)))
}

func (s *TelemetryService) HandleTelemetry(
	ctx context.Context,
	tm drone.Telemetry,
) error {
	lg := s.lg.With(
		zap.String("component", "telemetry_service"),
		zap.String("op", "HandleTelemetry"),
		zap.String("drone_id", tm.DroneID),
		zap.String("event", string(tm.Event)),
	)

	start := time.Now()

	if tm.DroneID == "" {
		lg.Warn(ctx, "validation failed: drone_id is empty")
		return fmt.Errorf("drone_id is required")
	}

	lg.Info(ctx, "telemetry received",
		zap.Float64("lat", tm.DroneLocation.Lat),
		zap.Float64("lon", tm.DroneLocation.Lon),
		zap.Float64("battery", tm.Battery),
		zap.Float64("speed_mps", tm.SpeedMps),
		zap.Float64("consumption_per_meter", tm.ConsumptionPerMeter),
		zap.String("status", string(tm.Status)),
		zap.Int64("timestamp", tm.Timestamp),
	)

	data := drone.TelemetryData{
		DroneID:             tm.DroneID,
		DroneLocation:       tm.DroneLocation,
		Battery:             tm.Battery,
		SpeedMps:            tm.SpeedMps,
		ConsumptionPerMeter: tm.ConsumptionPerMeter,
		Status:              tm.Status,
		Timestamp:           tm.Timestamp,
	}

	pubDataStart := time.Now()
	if err := resilience.Retry(ctx, telemetryRetryCfg, shouldRetryKafka, func(ctx context.Context) error {
		return s.publishData(ctx, data)
	}); err != nil {
		lg.Error(ctx, "failed to publish telemetry data after retries",
			zap.Duration("duration", time.Since(pubDataStart)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish telemetry data: %w", err)
	}
	lg.Info(ctx, "telemetry data published", zap.Duration("duration", time.Since(pubDataStart)))

	if tm.Event != drone.DroneEventNone {
		ev := drone.TelemetryEvent{
			DroneID:       tm.DroneID,
			DroneLocation: tm.DroneLocation,
			Event:         tm.Event,
			Timestamp:     tm.Timestamp,
		}

		pubEvStart := time.Now()
		if err := resilience.Retry(ctx, telemetryRetryCfg, shouldRetryKafka, func(ctx context.Context) error {
			return s.publishEvent(ctx, ev)
		}); err != nil {
			lg.Error(ctx, "failed to publish telemetry event after retries",
				zap.Duration("duration", time.Since(pubEvStart)),
				zap.Error(err),
			)
			return fmt.Errorf("failed to publish telemetry event: %w", err)
		}
		lg.Info(ctx, "telemetry event published", zap.Duration("duration", time.Since(pubEvStart)))
	}

	lg.Info(ctx, "handle telemetry completed", zap.Duration("duration", time.Since(start)))
	return nil
}

func (s *TelemetryService) EnqueueCommand(
	ctx context.Context,
	cmd *drone.ServerCommand,
) error {
	lg := s.lg.With(
		zap.String("component", "telemetry_service"),
		zap.String("op", "EnqueueCommand"),
	)

	start := time.Now()

	if cmd == nil {
		lg.Warn(ctx, "validation failed: cmd is nil")
		return fmt.Errorf("cmd is required")
	}
	if cmd.DroneID == "" {
		lg.Warn(ctx, "validation failed: drone_id is empty")
		return fmt.Errorf("drone_id is required")
	}

	lg = lg.With(
		zap.String("drone_id", cmd.DroneID),
		zap.String("action", string(cmd.Action)),
	)

	s.mu.RLock()
	conn, ok := s.conns[cmd.DroneID]
	s.mu.RUnlock()
	if !ok {
		lg.Warn(ctx, "drone not connected", zap.Duration("duration", time.Since(start)))
		return ErrDroneNotConnected
	}

	select {
	case conn.Commands <- cmd:
		lg.Info(ctx, "command enqueued",
			zap.Int("queue_len", len(conn.Commands)),
			zap.Int("queue_cap", cap(conn.Commands)),
			zap.Duration("duration", time.Since(start)),
		)
		return nil
	case <-ctx.Done():
		lg.Warn(ctx, "enqueue canceled by context", zap.Error(ctx.Err()), zap.Duration("duration", time.Since(start)))
		return ctx.Err()
	}
}

func (s *TelemetryService) publishEvent(
	ctx context.Context,
	event drone.TelemetryEvent,
) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.eventsWriter.WriteMessages(ctx, kafka.Message{
		Topic: s.eventsTopic,
		Value: bytes,
	})
}

func (s *TelemetryService) publishData(
	ctx context.Context,
	data drone.TelemetryData,
) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.dataWriter.WriteMessages(ctx, kafka.Message{
		Topic: s.dataTopic,
		Value: bytes,
	})
}

func shouldRetryKafka(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}
	return true
}
