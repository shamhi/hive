package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hive/services/telemetry/internal/domain/drone"
	"sync"

	"github.com/segmentio/kafka-go"
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
}

func NewTelemetryService(
	eventsWriter, dataWriter *kafka.Writer,
	eventsTopic, dataTopic string,
) *TelemetryService {
	return &TelemetryService{
		eventsWriter: eventsWriter,
		dataWriter:   dataWriter,
		eventsTopic:  eventsTopic,
		dataTopic:    dataTopic,
		conns:        make(map[string]*DroneConnection),
	}
}

func (s *TelemetryService) RegisterConnection(droneID string) *DroneConnection {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn := &DroneConnection{
		Commands: make(chan *drone.ServerCommand, 16),
	}
	s.conns[droneID] = conn
	return conn
}

func (s *TelemetryService) UnregisterConnection(droneID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.conns[droneID]; exists {
		close(conn.Commands)
		delete(s.conns, droneID)
	}
}

func (s *TelemetryService) HandleTelemetry(ctx context.Context, data drone.TelemetryData) error {
	if err := s.publishData(ctx, data); err != nil {
		return fmt.Errorf("failed to publish telemetry data: %w", err)
	}
	if data.Event != drone.DroneEventNone {
		if err := s.publishEvent(ctx, drone.TelemetryEvent{
			DroneID:       data.DroneID,
			DroneLocation: data.DroneLocation,
			Event:         data.Event,
			Timestamp:     data.Timestamp,
		}); err != nil {
			return fmt.Errorf("failed to publish telemetry event: %w", err)
		}
	}

	return nil
}

func (s *TelemetryService) EnqueueCommand(ctx context.Context, cmd *drone.ServerCommand) error {
	s.mu.RLock()
	conn, ok := s.conns[cmd.DroneID]
	s.mu.RUnlock()
	if !ok {
		return ErrDroneNotConnected
	}

	select {
	case conn.Commands <- cmd:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *TelemetryService) publishEvent(ctx context.Context, event drone.TelemetryEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.eventsWriter.WriteMessages(ctx, kafka.Message{
		Topic: s.eventsTopic,
		Value: bytes,
	})
}

func (s *TelemetryService) publishData(ctx context.Context, data drone.TelemetryData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.dataWriter.WriteMessages(ctx, kafka.Message{
		Topic: s.dataTopic,
		Value: bytes,
	})
}
