package service

import (
	"context"
	"errors"
	"fmt"
	"hive/pkg/logger"
	"hive/services/order/internal/domain/order"
	"hive/services/order/internal/domain/shared"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderService struct {
	repo     OrderRepository
	dispatch DispatchClient
	lg       logger.Logger
}

func NewOrderService(
	repo OrderRepository,
	dispatch DispatchClient,
	lg logger.Logger,
) *OrderService {
	return &OrderService{
		repo:     repo,
		dispatch: dispatch,
		lg:       lg,
	}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	userID string,
	items []string,
	deliveryLocation shared.Location,
) (*order.OrderInfo, error) {
	lg := s.lg.With(
		zap.String("component", "order_service"),
		zap.String("op", "CreateOrder"),
		zap.String("user_id", userID),
	)

	start := time.Now()

	if userID == "" {
		lg.Warn(ctx, "validation failed: user ID is empty")
		return nil, fmt.Errorf("user ID is required")
	}
	if len(items) == 0 {
		lg.Warn(ctx, "validation failed: items list is empty")
		return nil, fmt.Errorf("items list cannot be empty")
	}

	lg.Info(ctx, "create order started",
		zap.Int("items_count", len(items)),
		zap.Float64("delivery_lat", deliveryLocation.Lat),
		zap.Float64("delivery_lon", deliveryLocation.Lon),
	)

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
		lg.Error(ctx, "failed to save order", zap.String("order_id", o.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	lg = lg.With(zap.String("order_id", o.ID))
	lg.Info(ctx, "order saved", zap.String("status", string(o.Status)))

	lg.Info(ctx, "requesting drone assignment from dispatch")
	assignStart := time.Now()
	assign, err := s.dispatch.AssignDrone(ctx, o.ID, &o.Location)
	assignDur := time.Since(assignStart)
	if err != nil {
		lg.Warn(ctx, "dispatch assignment failed, returning PENDING (fallback)",
			zap.Duration("duration", assignDur),
			zap.Error(err),
		)
		lg.Info(ctx, "create order completed (fallback)",
			zap.String("final_status", string(order.OrderStatusPending)),
			zap.Duration("duration", time.Since(start)),
		)
		return &order.OrderInfo{
			ID:         o.ID,
			Status:     order.OrderStatusPending,
			DroneID:    "",
			EtaSeconds: 0,
		}, nil
	}

	lg.Info(ctx, "dispatch assignment succeeded",
		zap.String("drone_id", assign.DroneID),
		zap.Int32("eta_seconds", assign.EtaSeconds),
		zap.Duration("duration", assignDur),
	)

	if err := s.repo.UpdateDroneAndStatus(ctx, o.ID, assign.DroneID, order.OrderStatusAssigned); err != nil {
		lg.Error(ctx, "failed to update order with drone ID and status",
			zap.String("drone_id", assign.DroneID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update order with drone ID: %w", err)
	}

	lg.Info(ctx, "order marked as ASSIGNED",
		zap.String("drone_id", assign.DroneID),
		zap.String("final_status", string(order.OrderStatusAssigned)),
		zap.Duration("duration", time.Since(start)),
	)

	return &order.OrderInfo{
		ID:         o.ID,
		Status:     order.OrderStatusAssigned,
		DroneID:    assign.DroneID,
		EtaSeconds: assign.EtaSeconds,
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	lg := s.lg.With(
		zap.String("component", "order_service"),
		zap.String("op", "GetOrder"),
		zap.String("order_id", orderID),
	)

	start := time.Now()
	if orderID == "" {
		lg.Warn(ctx, "validation failed: order ID is empty")
		return nil, fmt.Errorf("order ID is required")
	}

	o, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			lg.Info(ctx, "order not found", zap.Duration("duration", time.Since(start)))
			return nil, ErrOrderNotFound
		}
		lg.Error(ctx, "failed to get order", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	lg.Info(ctx, "get order completed",
		zap.String("status", string(o.Status)),
		zap.String("user_id", o.UserID),
		zap.Int("items_count", len(o.Items)),
		zap.Duration("duration", time.Since(start)),
	)

	return o, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error {
	lg := s.lg.With(
		zap.String("component", "order_service"),
		zap.String("op", "UpdateStatus"),
		zap.String("order_id", orderID),
		zap.String("new_status", string(status)),
	)

	start := time.Now()
	if orderID == "" {
		lg.Warn(ctx, "validation failed: order ID is empty")
		return fmt.Errorf("order ID is required")
	}

	if err := s.repo.UpdateStatus(ctx, orderID, status); err != nil {
		lg.Error(ctx, "failed to update order status", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return fmt.Errorf("failed to update order status: %w", err)
	}

	lg.Info(ctx, "order status updated", zap.Duration("duration", time.Since(start)))
	return nil
}
