package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/domain/store"
)

type MockStoreClient struct{}

func NewMockStoreClient() *MockStoreClient {
	return &MockStoreClient{}
}

func (m *MockStoreClient) FindNearest(_ context.Context, deliveryLocation *shared.Location) (*store.StoreNearestInfo, error) {
	return nil, nil
}

func (m *MockStoreClient) GetStoreLocation(_ context.Context, storeID string) (*store.StoreInfo, error) {
	return nil, nil
}
