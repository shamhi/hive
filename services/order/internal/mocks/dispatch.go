package mocks

import (
	"context"
	"hive/services/order/internal/domain/shared"
)

type MockDispatchClient struct{}

func NewMockDispatchClient() *MockDispatchClient {
	return &MockDispatchClient{}
}

func (m *MockDispatchClient) AssignDrone(_ context.Context, orderID string, location *shared.Location) (string, error) {
	return "drone-mock-123", nil
}
