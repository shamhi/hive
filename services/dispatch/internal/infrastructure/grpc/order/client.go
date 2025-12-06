package order

import (
	"context"
	"fmt"
	pbOrder "hive/gen/order"
	"hive/services/dispatch/internal/domain/mapping"
	"hive/services/dispatch/internal/domain/order"
)

type OrderClient struct {
	client pbOrder.OrderServiceClient
}

func NewOrderClient(client pbOrder.OrderServiceClient) *OrderClient {
	return &OrderClient{client: client}
}

func (c *OrderClient) UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error {
	req := &pbOrder.UpdateStatusRequest{
		OrderId: orderID,
		Status:  mapping.OrderStatusToProto(status),
	}
	resp, err := c.client.UpdateStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("order status update was not successful")
	}

	return nil
}
