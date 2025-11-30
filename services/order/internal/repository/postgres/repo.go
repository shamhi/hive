package postgres

import (
	"context"
	"hive/services/order/internal/domain"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db      *pgx.Pool
	builder sq.StatementBuilderType
}

func NewPostgresRepo(db *pgx.Pool) *PostgresRepo {
	return &PostgresRepo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PostgresRepo) Save(ctx context.Context, o *domain.Order) error {
	sql, args, err := sq.
		Insert("orders").
		Columns("id", "items", "status").
		Values(o.ID, o.Items, o.Status).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepo) Get(ctx context.Context, id string) (*domain.Order, error) {
	sql, args, err := sq.
		Select("id", "items", "status").
		From("orders").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var o domain.Order
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&o.ID, &o.Items, &o.Status); err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *PostgresRepo) Update(ctx context.Context, o *domain.Order) error {
	sql, args, err := sq.
		Update("orders").
		Set("items", o.Items).
		Set("status", o.Status).
		Set("location", o.Location).
		Where(sq.Eq{"id": o.ID}).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}
