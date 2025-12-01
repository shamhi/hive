package postgres

import "fmt"

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
