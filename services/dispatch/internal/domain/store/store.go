package store

import "hive/services/dispatch/internal/domain/shared"

type Store struct {
	ID       string
	Location shared.Location
}

type StoreNearest struct {
	ID       string
	Distance float64
}
