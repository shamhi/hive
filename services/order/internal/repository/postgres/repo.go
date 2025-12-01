package postgres

import (
	"context"
	"errors"
	"hive/services/order/internal/domain"
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

func (r *PostgresRepo) Save(ctx context.Context, o *domain.Order) error {
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
			string(o.Status),
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

func (r *PostgresRepo) Get(ctx context.Context, id string) (*domain.Order, error) {
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

	var (
		o         domain.Order
		statusStr string
	)

	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&o.ID,
		&o.UserID,
		&o.DroneID,
		&o.Items,
		&statusStr,
		&o.Location.Lat,
		&o.Location.Lon,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrNotFound
		}
		return nil, err
	}

	o.Status = domain.OrderStatus(statusStr)
	return &o, nil
}

func (r *PostgresRepo) Update(ctx context.Context, o *domain.Order) error {
	sql, args, err := r.builder.
		Update("orders").
		Set("drone_id", o.DroneID).
		Set("items", o.Items).
		Set("status", string(o.Status)).
		Set("delivery_lat", o.Location.Lat).
		Set("delivery_lon", o.Location.Lon).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": o.ID}).
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
