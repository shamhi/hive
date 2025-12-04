package postgres

import (
	"context"
	"errors"
	"fmt"
	"hive/services/order/internal/domain/order"
	"hive/services/order/internal/service"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db      *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PostgresRepo) Save(ctx context.Context, o *order.Order) error {
	sql, args, err := r.builder.
		Insert("orders").
		Columns(
			"id",
			"user_id",
			"drone_id",
			"items",
			"status",
			"delivery_lat",
			"delivery_lon",
			"created_at",
			"updated_at",
		).
		Values(
			o.ID,
			o.UserID,
			o.DroneID,
			o.Items,
			o.Status,
			o.Location.Lat,
			o.Location.Lon,
			o.CreatedAt,
			o.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (*order.Order, error) {
	sql, args, err := r.builder.
		Select(
			"id",
			"user_id",
			"drone_id",
			"items",
			"status",
			"delivery_lat",
			"delivery_lon",
			"created_at",
			"updated_at",
		).
		From("orders").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var o order.Order
	if err = r.db.QueryRow(ctx, sql, args...).Scan(
		&o.ID,
		&o.UserID,
		&o.DroneID,
		&o.Items,
		&o.Status,
		&o.Location.Lat,
		&o.Location.Lon,
		&o.CreatedAt,
		&o.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrNotFound
		}
		return nil, err
	}

	return &o, nil
}

func (r *PostgresRepo) UpdateStatus(ctx context.Context, id string, status order.OrderStatus) error {
	sql, args, err := r.builder.
		Update("orders").
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return service.ErrNotFound
	}

	return nil
}

func (r *PostgresRepo) SetDroneID(ctx context.Context, id string, droneID string) error {
	sql, args, err := r.builder.
		Update("orders").
		Set("drone_id", droneID).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return service.ErrNotFound
	}

	return nil
}

func (r *PostgresRepo) UpdateDroneAndStatus(ctx context.Context, id string, droneID string, status order.OrderStatus) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sql, args, err := r.builder.
		Update("orders").
		Set("drone_id", droneID).
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return service.ErrNotFound
	}

	return tx.Commit(ctx)
}
