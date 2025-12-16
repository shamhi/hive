package store

import (
	"hive/services/api-gateway/internal/domain/shared"
)

type Store struct {
	ID string

	Name     string
	Address  string
	Location shared.Location
}

type StoreNearest struct {
	ID       string
	Distance float64
}
