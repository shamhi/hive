package base

import "hive/services/dispatch/internal/domain/shared"

type Base struct {
	ID       string
	Name     string
	Address  string
	Location shared.Location
}

type BaseNearest struct {
	ID       string
	Distance float64
}
