package client

import (
	"context"
	"fmt"
	"hive/gen/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	client order.OrderClient
	conn   *grpc.ClientConn
}

func NewOrderClient(addr string) (*OrderClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &OrderClient{
		client: order.NewOrderServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *OrderClient) Close() error {
	return c.conn.Close()
}

func (c *OrderClient) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	return c.client.CreateOrder(ctx, req)
}

func (c *OrderClient) GetOrder(ctx context.Context, req *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	return c.client.GetOrder(ctx, req)
}
