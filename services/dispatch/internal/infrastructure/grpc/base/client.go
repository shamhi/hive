package base

import (
	"context"
	"fmt"
	pbBase "hive/gen/base"
	pbCommon "hive/gen/common"
	"hive/services/dispatch/internal/domain/base"
	"hive/services/dispatch/internal/domain/shared"
)

type BaseClient struct {
	client pbBase.BaseServiceClient
}

func NewBaseClient(client pbBase.BaseServiceClient) *BaseClient {
	return &BaseClient{client: client}
}

func (c *BaseClient) FindNearest(ctx context.Context, deliveryLocation *shared.Location) (*base.BaseNearest, error) {
	req := &pbBase.FindNearestRequest{
		DroneLocation: &pbCommon.Location{
			Lat: deliveryLocation.Lat,
			Lon: deliveryLocation.Lon,
		},
	}
	resp, err := c.client.FindNearest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest base: %w", err)
	}

	if !resp.GetFound() {
		return nil, fmt.Errorf("no base found")
	}

	if resp.GetBaseId() == "" {
		return nil, fmt.Errorf("no base found")
	}
	if resp.GetDistanceMeters() <= 0 {
		return nil, fmt.Errorf("invalid distance returned for nearest base")
	}

	return &base.BaseNearest{
		ID:       resp.GetBaseId(),
		Distance: resp.GetDistanceMeters(),
	}, nil
}

func (c *BaseClient) GetBaseLocation(ctx context.Context, baseID string) (*base.Base, error) {
	req := &pbBase.GetBaseLocationRequest{
		BaseId: baseID,
	}
	resp, err := c.client.GetBaseLocation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get base location: %w", err)
	}

	if resp.GetLocation() == nil {
		return nil, fmt.Errorf("base location not found")
	}

	return &base.Base{
		ID:      baseID,
		Name:    resp.GetName(),
		Address: resp.GetAddress(),
		Location: shared.Location{
			Lat: resp.GetLocation().GetLat(),
			Lon: resp.GetLocation().GetLon(),
		},
	}, nil
}
