package service

import (
	"context"
	"fmt"
	"hive/services/order/internal/domain/order"
	"hive/services/order/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type OrderService struct {
	repo     OrderRepository
	dispatch DispatchClient
}

func NewOrderService(
	repo OrderRepository,
	dispatch DispatchClient,
) *OrderService {
	return &OrderService{
		repo:     repo,
		dispatch: dispatch,
	}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	userID string,
	items []string,
	deliveryLocation shared.Location,
) (*order.OrderInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("items list cannot be empty")
	}
	if deliveryLocation.Lat == 0 && deliveryLocation.Lon == 0 {
		return nil, fmt.Errorf("delivery location is required")
	}

	o := &order.Order{
		ID:        uuid.NewString(),
		UserID:    userID,
		Items:     items,
		Status:    order.OrderStatusPending,
		Location:  deliveryLocation,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, o); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	droneID, err := s.dispatch.AssignDrone(ctx, o.ID, &o.Location)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, o.ID, order.OrderStatusFailed)
		return nil, fmt.Errorf("assign drone failed: %w", err)
	}

	if err := s.repo.UpdateDroneAndStatus(ctx, o.ID, droneID, order.OrderStatusAssigned); err != nil {
		return nil, fmt.Errorf("failed to update order with drone ID: %w", err)
	}

	return &order.OrderInfo{
		ID:         o.ID,
		Status:     order.OrderStatusAssigned,
		DroneID:    droneID,
		EtaSeconds: 900, // TODO: calculate ETA
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	o, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return o, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error {
	if err := s.repo.UpdateStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}
