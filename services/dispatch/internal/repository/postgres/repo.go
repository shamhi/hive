package postgres

import (
	"context"
	"errors"
	"hive/services/dispatch/internal/domain"
	"hive/services/dispatch/internal/service"

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

func (r *PostgresRepo) Save(ctx context.Context, assignment *domain.Assignment) error {
	sql, args, err := r.builder.
		Insert("assignments").
		Columns(
			"id",
			"order_id",
			"drone_id",
			"status",
			"created_at",
			"updated_at",
		).
		Values(
			assignment.ID,
			assignment.OrderID,
			assignment.DroneID,
			assignment.Status,
			assignment.CreatedAt,
			assignment.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepo) Get(ctx context.Context, id string) (*domain.Assignment, error) {
	sql, args, err := r.builder.
		Select(
			"id",
			"order_id",
			"drone_id",
			"status",
			"created_at",
			"updated_at",
		).
		From("assignments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var assignment domain.Assignment
	if err := r.db.QueryRow(ctx, sql, args...).Scan(
		&assignment.ID,
		&assignment.OrderID,
		&assignment.DroneID,
		&assignment.Status,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrNotFound
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *PostgresRepo) Update(ctx context.Context, assignment *domain.Assignment) error {
	sql, args, err := r.builder.
		Update("assignments").
		Set("order_id", assignment.OrderID).
		Set("drone_id", assignment.DroneID).
		Set("status", assignment.Status).
		Set("updated_at", assignment.UpdatedAt).
		Where(sq.Eq{"id": assignment.ID}).
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
