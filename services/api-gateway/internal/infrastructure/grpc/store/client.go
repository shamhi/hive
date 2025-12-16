package store

import (
	"context"
	pbStore "hive/gen/store"
	"hive/services/api-gateway/internal/domain/store"
)

type StoreClient struct {
	client pbStore.StoreServiceClient
}

func NewStoreClient(client pbStore.StoreServiceClient) *StoreClient {
	return &StoreClient{
		client: client,
	}
}

func (c *StoreClient) ListStores(
	ctx context.Context,
	offset, limit int64,
) ([]*store.Store, error) {
	return []*store.Store{}, nil
}
