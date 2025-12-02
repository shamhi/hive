package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain"
)

type MockOrderClient struct{}

func NewMockOrderClient() *MockOrderClient {
	return &MockOrderClient{}
}

func (m *MockOrderClient) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	return nil
}
