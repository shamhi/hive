package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain"
)

type MockTelemetryClient struct{}

func NewMockTelemetryClient() *MockTelemetryClient {
	return &MockTelemetryClient{}
}

func (m *MockTelemetryClient) SendCommand(
	_ context.Context,
	droneID string,
	action domain.DroneAction,
	target domain.Location,
) error {
	return nil
}
