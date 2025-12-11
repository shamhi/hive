package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pg "hive/pkg/db/postgres"
)

func main() {
	cmd := flag.String("command", "up", "migration command: up | down | version")
	migrationsPath := flag.String("path", "migrations", "path to migrations folder")
	flag.Parse()

	var cfg pg.Config
	if err := env.Parse(&cfg); err != nil {
		fatal("failed to parse env config", err)
	}

	dsn := cfg.DSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fatal("failed to open DB", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fatal("failed to create migration driver", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		fatal("failed to load migrations", err)
	}

	switch *cmd {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fatal("migration up failed", err)
		}
		fmt.Println("migrations applied")

	case "down":
		if err := m.Down(); err != nil {
			fatal("migration down failed", err)
		}
		fmt.Println("migrations reverted")

	case "version":
		v, dirty, err := m.Version()
		if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
			fatal("failed to get version", err)
		}
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("version: nil")
		} else {
			fmt.Printf("version: %d (dirty: %v)\n", v, dirty)
		}

	default:
		fatal("unknown command", *cmd)
	}
}

func fatal(msg string, val any) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, val)
	os.Exit(1)
}
