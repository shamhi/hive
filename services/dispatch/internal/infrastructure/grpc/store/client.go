package store

import (
	"context"
	"fmt"
	pbCommon "hive/gen/common"
	pbStore "hive/gen/store"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/domain/store"
)

type StoreClient struct {
	client pbStore.StoreServiceClient
}

func NewStoreClient(client pbStore.StoreServiceClient) *StoreClient {
	return &StoreClient{client: client}
}

func (c *StoreClient) FindNearest(ctx context.Context, deliveryLocation *shared.Location) (*store.StoreNearest, error) {
	req := &pbStore.FindNearestRequest{
		DeliveryLocation: &pbCommon.Location{
			Lat: deliveryLocation.Lat,
			Lon: deliveryLocation.Lon,
		},
	}
	resp, err := c.client.FindNearest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest store: %w", err)
	}

	if !resp.GetFound() {
		return nil, fmt.Errorf("no store found")
	}
	if resp.GetStoreId() == "" {
		return nil, fmt.Errorf("no store found")
	}
	if resp.GetDistanceMeters() < 0 {
		return nil, fmt.Errorf("invalid distance returned for nearest store")
	}

	return &store.StoreNearest{
		ID:       resp.GetStoreId(),
		Distance: resp.GetDistanceMeters(),
	}, nil
}

func (c *StoreClient) GetStoreLocation(ctx context.Context, storeID string) (*store.Store, error) {
	req := &pbStore.GetStoreLocationRequest{
		StoreId: storeID,
	}
	resp, err := c.client.GetStoreLocation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get store location: %w", err)
	}

	if resp.GetLocation() == nil {
		return nil, fmt.Errorf("store location not found")
	}

	return &store.Store{
		ID:      storeID,
		Name:    resp.GetName(),
		Address: resp.GetAddress(),
		Location: shared.Location{
			Lat: resp.GetLocation().GetLat(),
			Lon: resp.GetLocation().GetLon(),
		},
	}, nil
}
