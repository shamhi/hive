package service

import (
	"context"
	"fmt"
	"hive/services/order/internal/domain"
)

type OrderService struct {
	repo       OrderRepository
	dispatcher DispatchClient
}

func NewOrderService(repo OrderRepository, dispatcher DispatchClient) *OrderService {
	return &OrderService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, id string, items []string, loc domain.Location) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("empty items list")
	}

	order := &domain.Order{
		ID:       id,
		Items:    items,
		Status:   domain.CREATED,
		Location: loc,
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return "", err
	}

	if err := s.dispatcher.AssignDrone(ctx, order.ID, order.Location); err != nil {
		order.Status = domain.FAILED
		if err := s.repo.Update(ctx, order); err != nil {
			return "", err
		}
		return "", err
	}

	order.Status = domain.PENDING
	if err := s.repo.Update(ctx, order); err != nil {
		return "", nil
	}

	// TODO: calculate estimatedTime

	return "15min", nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return order, nil
}
