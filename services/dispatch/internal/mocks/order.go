package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain/order"
)

type MockOrderClient struct{}

func NewMockOrderClient() *MockOrderClient {
	return &MockOrderClient{}
}

func (m *MockOrderClient) UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error {
	return nil
}
