package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/shared"
)

type MockTrackingClient struct{}

func NewMockTrackingClient() *MockTrackingClient {
	return &MockTrackingClient{}
}

func (m *MockTrackingClient) FindNearest(
	_ context.Context,
	storeLocation *shared.Location,
	minBattery float64,
	radius float64,
) (*drone.DroneNearest, error) {
	return nil, nil
}

func (m *MockTrackingClient) GetDroneLocation(_ context.Context, droneID string) (*drone.Drone, error) {
	return nil, nil
}

func (m *MockTrackingClient) SetStatus(_ context.Context, droneID string, status drone.DroneStatus) error {
	return nil
}
