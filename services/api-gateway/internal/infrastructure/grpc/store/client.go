package store

import (
	"context"
	"fmt"
	pbStore "hive/gen/store"
	"hive/services/api-gateway/internal/domain/mapping"
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
	req := &pbStore.ListStoresRequest{
		Offset: offset,
		Limit:  limit,
	}
	resp, err := c.client.ListStores(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListStores: %w", err)
	}
	pbStores := resp.GetStores()
	stores := make([]*store.Store, 0, len(pbStores))
	for _, pbS := range resp.GetStores() {
		s, ok := mapping.StoreFromProto(pbS)
		if !ok {
			continue
		}
		stores = append(stores, s)
	}
	return stores, nil
}
