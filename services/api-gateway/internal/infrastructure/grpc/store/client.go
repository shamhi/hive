package store

import (
	"context"
	pbStore "hive/gen/store"
	"hive/services/api-gateway/internal/domain/mapping"
	"hive/services/api-gateway/internal/domain/store"
)

type StoreClient struct {
	client pbStore.StoreServiceClient
}

func NewStoreClient(client pbStore.StoreServiceClient) *StoreClient {
	return &StoreClient{client: client}
}

func (c *StoreClient) ListStores(ctx context.Context, offset, limit int64) ([]*store.Store, error) {
	req := &pbStore.ListStoresRequest{
		Offset: offset,
		Limit:  limit,
	}

	resp, err := c.client.ListStores(ctx, req)
	if err != nil {
		return nil, err
	}

	pbStores := resp.GetStores()
	out := make([]*store.Store, 0, len(pbStores))
	for _, pbS := range pbStores {
		s, ok := mapping.StoreFromProto(pbS)
		if !ok || s == nil {
			continue
		}
		out = append(out, s)
	}

	return out, nil
}
