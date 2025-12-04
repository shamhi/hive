package order

import (
	"context"
	"fmt"
	pbOrder "hive/gen/order"
	"hive/services/dispatch/internal/domain/order"
)

type OrderClient struct {
	client pbOrder.OrderServiceClient
}

func NewOrderClient(client pbOrder.OrderServiceClient) *OrderClient {
	return &OrderClient{client: client}
}

func (c *OrderClient) UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error {
	var newStatus pbOrder.OrderStatus
	switch status {
	case order.OrderStatusPending:
		newStatus = pbOrder.OrderStatus_PENDING
	case order.OrderStatusAssigned:
		newStatus = pbOrder.OrderStatus_ASSIGNED
	case order.OrderStatusCompleted:
		newStatus = pbOrder.OrderStatus_COMPLETED
	case order.OrderStatusFailed:
		newStatus = pbOrder.OrderStatus_FAILED
	default:
		return fmt.Errorf("invalid order status: %v", status)
	}

	req := &pbOrder.UpdateStatusRequest{
		OrderId: orderID,
		Status:  newStatus,
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
