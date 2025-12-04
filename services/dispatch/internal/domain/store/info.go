package store

import "hive/services/dispatch/internal/domain/shared"

type StoreInfo struct {
	ID       string
	Location shared.Location
}

type StoreNearestInfo struct {
	ID       string
	Distance float64
}
