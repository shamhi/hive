package store

import (
	"hive/services/store/internal/domain/shared"
	"time"
)

type Store struct {
	ID string

	Name     string
	Address  string
	Location shared.Location

	CreatedAt time.Time
	UpdatedAt time.Time
}

type StoreNearest struct {
	ID       string
	Distance float64
}
