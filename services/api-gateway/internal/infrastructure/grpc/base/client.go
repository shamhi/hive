package base

import (
	"context"
	"fmt"

	pbBase "hive/gen/base"
	"hive/services/api-gateway/internal/domain/base"
	"hive/services/api-gateway/internal/domain/mapping"
)

type BaseClient struct {
	client pbBase.BaseServiceClient
}

func NewBaseClient(client pbBase.BaseServiceClient) *BaseClient {
	return &BaseClient{client: client}
}

func (c *BaseClient) ListBases(ctx context.Context, offset, limit int64) ([]*base.Base, error) {
	req := &pbBase.ListBasesRequest{
		Offset: offset,
		Limit:  limit,
	}

	resp, err := c.client.ListBases(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list bases: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("received nil response when listing bases")
	}

	pbBases := resp.GetBases()
	out := make([]*base.Base, 0, len(pbBases))
	for _, pbB := range pbBases {
		b, ok := mapping.BaseFromProto(pbB)
		if !ok || b == nil {
			continue
		}
		out = append(out, b)
	}

	return out, nil
}
