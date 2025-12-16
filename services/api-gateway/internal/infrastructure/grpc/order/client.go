package order

import (
	"context"
	"fmt"
	pbCommon "hive/gen/common"
	pbOrder "hive/gen/order"
	"hive/services/api-gateway/internal/domain/mapping"
	"hive/services/api-gateway/internal/domain/order"
	"hive/services/api-gateway/internal/domain/shared"
)

type OrderClient struct {
	client pbOrder.OrderServiceClient
}

func NewOrderClient(client pbOrder.OrderServiceClient) *OrderClient {
	return &OrderClient{
		client: client,
	}
}

func (c *OrderClient) CreateOrder(
	ctx context.Context,
	userID string,
	items []string,
	location shared.Location,
) (*order.OrderInfo, error) {
	req := &pbOrder.CreateOrderRequest{
		UserId: userID,
		Items:  items,
		DeliveryLocation: &pbCommon.Location{
			Lat: location.Lat,
			Lon: location.Lon,
		},
	}
	resp, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	st, ok := mapping.OrderStatusFromProto(resp.GetStatus())
	if !ok {
		return nil, fmt.Errorf("unknown order status from proto: %v", resp.GetStatus())
	}

	return &order.OrderInfo{
		ID:         resp.GetOrderId(),
		DroneID:    resp.GetDroneId(),
		Status:     st,
		EtaSeconds: resp.GetEtaSeconds(),
	}, nil
}

func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	req := &pbOrder.GetOrderRequest{
		OrderId: orderID,
	}
	resp, err := c.client.GetOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	st, ok := mapping.OrderStatusFromProto(resp.GetStatus())
	if !ok {
		return nil, fmt.Errorf("unknown order status from proto: %v", resp.GetStatus())
	}

	return &order.Order{
		ID:      resp.GetOrderId(),
		UserID:  resp.GetUserId(),
		DroneID: resp.GetDroneId(),
		Items:   resp.GetItems(),
		Status:  st,
		Location: shared.Location{
			Lat: resp.GetDeliveryLocation().GetLat(),
			Lon: resp.GetDeliveryLocation().GetLon(),
		},
	}, nil
}
