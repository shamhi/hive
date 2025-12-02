package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain"
)

type MockTrackingClient struct{}

func NewMockTrackingClient() *MockTrackingClient {
	return &MockTrackingClient{}
}

func (m *MockTrackingClient) FindNearest(_ context.Context, storeLocation domain.Location) (string, error) {
	return "mock-drone-123", nil
}

func (m *MockTrackingClient) SetStatus(_ context.Context, droneID string, status domain.DroneStatus) error {
	return nil
}
