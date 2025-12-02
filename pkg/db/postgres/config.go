package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Username string `env:"POSTGRES_USERNAME" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Host     string `env:"POSTGRES_HOST"     env-default:"postgres"`
	Port     int    `env:"POSTGRES_PORT"     env-default:"5432"`
	DBName   string `env:"POSTGRES_DB"       env-default:"postgres"`
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}

type Database struct {
	Pool *pgxpool.Pool
}

func New(cfg Config) (*Database, error) {
	dsn := cfg.DSN()

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx.New: %w", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgx.Pool.Ping: %w", err)
	}

	return &Database{Pool: pool}, nil
}

func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}
