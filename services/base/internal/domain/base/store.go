package base

import (
	"hive/services/base/internal/domain/shared"
	"time"
)

type Base struct {
	ID string

	Name     string
	Address  string
	Location shared.Location

	CreatedAt time.Time
	UpdatedAt time.Time
}

type BaseNearest struct {
	ID       string
	Distance float64
}
