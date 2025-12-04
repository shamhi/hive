package client

import (
	"context"
	"fmt"
	"hive/gen/tracking"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TrackingClient struct {
	client tracking.TrackingServiceClient
	conn   *grpc.ClientConn
}

func NewTrackingClient(addr string) (*TrackingClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tracking service: %w", err)
	}

	return &TrackingClient{
		client: tracking.NewTrackingServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *TrackingClient) Close() error {
	return c.conn.Close()
}

func (c *TrackingClient) GetDroneLocation(ctx context.Context, req *tracking.GetDroneLocationRequest) (*tracking.GetDroneLocationResponse, error) {
	return c.client.GetDroneLocation(ctx, req)
}
