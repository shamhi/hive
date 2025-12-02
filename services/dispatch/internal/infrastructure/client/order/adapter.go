package order

import (
	"context"
	"fmt"
	pbOrder "hive/gen/order"
	"hive/services/dispatch/internal/domain"
)

type OrderAdapter struct {
	client pbOrder.OrderServiceClient
}

func NewOrderAdapter(client pbOrder.OrderServiceClient) *OrderAdapter {
	return &OrderAdapter{client: client}
}

func (a *OrderAdapter) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	var newStatus pbOrder.OrderStatus
	switch status {
	case domain.OrderStatusPending:
		newStatus = pbOrder.OrderStatus_PENDING
	case domain.OrderStatusAssigned:
		newStatus = pbOrder.OrderStatus_ASSIGNED
	case domain.OrderStatusCompleted:
		newStatus = pbOrder.OrderStatus_COMPLETED
	case domain.OrderStatusFailed:
		newStatus = pbOrder.OrderStatus_FAILED
	default:
		newStatus = pbOrder.OrderStatus_CREATED
	}

	req := &pbOrder.UpdateStatusRequest{
		OrderId: orderID,
		Status:  newStatus,
	}
	resp, err := a.client.UpdateStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("order status update was not successful")
	}

	return nil
}
