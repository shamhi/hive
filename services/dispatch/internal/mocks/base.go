package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain/base"
	"hive/services/dispatch/internal/domain/shared"
)

type MockBaseClient struct{}

func NewMockBaseClient() *MockBaseClient {
	return &MockBaseClient{}
}

func (m *MockBaseClient) FindNearest(_ context.Context, deliveryLocation *shared.Location) (*base.BaseNearest, error) {
	return nil, nil
}

func (m *MockBaseClient) GetBaseLocation(_ context.Context, baseID string) (*base.Base, error) {
	return nil, nil
}
