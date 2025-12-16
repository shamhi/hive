package base

import (
	"context"
	pbBase "hive/gen/base"
	"hive/services/api-gateway/internal/domain/base"
)

type BaseClient struct {
	client pbBase.BaseServiceClient
}

func NewBaseClient(client pbBase.BaseServiceClient) *BaseClient {
	return &BaseClient{
		client: client,
	}
}

func (c *BaseClient) ListBases(
	ctx context.Context,
	offset, limit int64,
) ([]*base.Base, error) {
	return []*base.Base{}, nil
}
