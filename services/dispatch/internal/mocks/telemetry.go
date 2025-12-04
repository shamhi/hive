package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain/drone"
)

type MockTelemetryClient struct{}

func NewMockTelemetryClient() *MockTelemetryClient {
	return &MockTelemetryClient{}
}

func (m *MockTelemetryClient) SendCommand(
	_ context.Context,
	droneID string,
	action drone.DroneAction,
	target *drone.Target,
) error {
	return nil
}
