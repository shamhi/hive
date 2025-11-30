package mock

import (
	"context"
	"hive/services/order/internal/domain"
)

type MockDispatch struct{}

func NewMockDispatch() *MockDispatch {
	return &MockDispatch{}
}

func (m *MockDispatch) AssignDrone(_ context.Context, orderID string, loc domain.Location) error {
	return nil
}
