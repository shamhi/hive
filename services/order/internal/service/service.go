package service

import (
	"context"
	"fmt"
	"hive/services/order/internal/domain"
	"time"
)

type OrderService struct {
	repo       OrderRepository
	dispatcher DispatchClient
}

func NewOrderService(repo OrderRepository, dispatcher DispatchClient) *OrderService {
	return &OrderService{repo: repo, dispatcher: dispatcher}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	orderID, userID string,
	items []string,
	loc domain.Location,
) (droneID string, etaSeconds int32, err error) {

	order := &domain.Order{
		ID:        orderID,
		UserID:    userID,
		Items:     items,
		Status:    domain.PENDING,
		Location:  loc,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return "", 0, fmt.Errorf("failed to save order: %w", err)
	}

	droneID, err = s.dispatcher.AssignDrone(ctx, order.ID, order.Location)
	if err != nil {
		order.Status = domain.FAILED
		_ = s.repo.Update(ctx, order)
		return "", 0, fmt.Errorf("assign drone failed: %w", err)
	}

	order.DroneID = droneID
	order.Status = domain.ASSIGNED
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return "", 0, fmt.Errorf("failed to update order: %w", err)
	}

	return droneID, 900, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	order, err := s.repo.Get(ctx, orderID)
	if err != nil {
		return err
	}

	order.Status = status
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}
