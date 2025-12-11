package repository

import (
	"context"
	"fmt"
	"hive/services/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (id, user_id, items, delivery_lat, delivery_lon, status, drone_id, estimated_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		order.ID,
		order.UserID,
		order.Items,
		order.DeliveryLat,
		order.DeliveryLon,
		order.Status,
		order.DroneID,
		order.EstimatedTime,
		order.CreatedAt,
		order.UpdatedAt,
	)

	return err
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	query := `
		SELECT id, user_id, items, delivery_lat, delivery_lon, status, drone_id, estimated_time, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Items,
		&order.DeliveryLat,
		&order.DeliveryLon,
		&order.Status,
		&order.DroneID,
		&order.EstimatedTime,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}
	order.SetDeliveryLocation()
	return &order, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE orders 
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, items, delivery_lat, delivery_lon, status, drone_id, estimated_time, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Items,
			&order.DeliveryLat,
			&order.DeliveryLon,
			&order.Status,
			&order.DroneID,
			&order.EstimatedTime,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		order.SetDeliveryLocation()
		orders = append(orders, &order)
	}

	return orders, nil
}
