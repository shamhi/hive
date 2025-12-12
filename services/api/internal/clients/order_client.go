package clients

import (
	"context"
	"fmt"
	"hive/gen/common"
	"time"

	"hive/gen/order"
	pb "hive/gen/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	client pb.OrderServiceClient
	conn   *grpc.ClientConn
}

func NewOrderClient(address string, timeout time.Duration) (*OrderClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	client := pb.NewOrderServiceClient(conn)

	return &OrderClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *OrderClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *OrderClient) CreateOrder(ctx context.Context, userID string, items []string, lat, lon float64) (*order.CreateOrderResponse, error) {
	req := &order.CreateOrderRequest{
		UserId: userID,
		Items:  items,
		DeliveryLocation: &common.Location{
			Lat: lat,
			Lon: lon,
		},
	}

	resp, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return resp, nil
}

func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*order.GetOrderResponse, error) {
	req := &order.GetOrderRequest{
		OrderId: orderID,
	}

	resp, err := c.client.GetOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return resp, nil
}

func (c *OrderClient) UpdateStatus(ctx context.Context, orderID string, status pb.OrderStatus) (bool, error) {
	req := &order.UpdateStatusRequest{
		OrderId: orderID,
		Status:  status,
	}

	resp, err := c.client.UpdateStatus(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to update status: %w", err)
	}

	return resp.Success, nil
}

func (c *OrderClient) UpdateStatusString(ctx context.Context, orderID string, status string) (bool, error) {
	var orderStatus pb.OrderStatus

	switch status {
	case "PENDING":
		orderStatus = pb.OrderStatus_PENDING
	case "ASSIGNED":
		orderStatus = pb.OrderStatus_ASSIGNED
	case "COMPLETED":
		orderStatus = pb.OrderStatus_COMPLETED
	case "FAILED":
		orderStatus = pb.OrderStatus_FAILED
	default:
		orderStatus = pb.OrderStatus_PENDING
	}

	req := &order.UpdateStatusRequest{
		OrderId: orderID,
		Status:  orderStatus,
	}

	resp, err := c.client.UpdateStatus(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to update status: %w", err)
	}

	return resp.Success, nil
}
