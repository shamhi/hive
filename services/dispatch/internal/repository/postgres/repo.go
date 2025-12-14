package postgres

import (
	"context"
	"errors"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/service"
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

func (r *PostgresRepo) Save(ctx context.Context, a *assignment.Assignment) error {
	var lat, lon *float64
	if a.Target != nil {
		lat = &a.Target.Lat
		lon = &a.Target.Lon
	}

	sql, args, err := r.builder.
		Insert("assignments").
		Columns(
			"id",
			"order_id",
			"drone_id",
			"status",
			"target_lat",
			"target_lon",
			"created_at",
			"updated_at",
		).
		Values(
			a.ID,
			a.OrderID,
			a.DroneID,
			a.Status,
			lat,
			lon,
			a.CreatedAt,
			a.UpdatedAt,
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

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (*assignment.Assignment, error) {
	sql, args, err := r.builder.
		Select(
			"id",
			"order_id",
			"drone_id",
			"status",
			"target_lat",
			"target_lon",
			"created_at",
			"updated_at",
		).
		From("assignments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var a assignment.Assignment
	var targetLat, targetLon *float64
	if err := r.db.QueryRow(ctx, sql, args...).Scan(
		&a.ID,
		&a.OrderID,
		&a.DroneID,
		&a.Status,
		&targetLat,
		&targetLon,
		&a.CreatedAt,
		&a.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrAssignmentNotFound
		}
		return nil, err
	}

	if targetLat != nil && targetLon != nil {
		a.Target = &shared.Location{
			Lat: *targetLat,
			Lon: *targetLon,
		}
	}

	return &a, nil
}

func (r *PostgresRepo) GetByDroneID(ctx context.Context, droneID string) (*assignment.Assignment, error) {
	sql, args, err := r.builder.
		Select(
			"id",
			"order_id",
			"drone_id",
			"status",
			"target_lat",
			"target_lon",
			"created_at",
			"updated_at",
		).
		From("assignments").
		Where(sq.Eq{"drone_id": droneID}).
		Where(sq.NotEq{"status": assignment.AssignmentStatusCompleted}).
		Where(sq.NotEq{"status": assignment.AssignmentStatusFailed}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var a assignment.Assignment
	var targetLat, targetLon *float64
	if err := r.db.QueryRow(ctx, sql, args...).Scan(
		&a.ID,
		&a.OrderID,
		&a.DroneID,
		&a.Status,
		&targetLat,
		&targetLon,
		&a.CreatedAt,
		&a.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrAssignmentNotFound
		}
		return nil, err
	}

	if targetLat != nil && targetLon != nil {
		a.Target = &shared.Location{
			Lat: *targetLat,
			Lon: *targetLon,
		}
	}

	return &a, nil
}

func (r *PostgresRepo) UpdateStatus(ctx context.Context, id string, status assignment.AssignmentStatus) error {
	sql, args, err := r.builder.
		Update("assignments").
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
		return service.ErrAssignmentNotFound
	}

	return nil
}
