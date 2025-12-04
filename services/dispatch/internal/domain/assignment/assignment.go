package assignment

import (
	"hive/services/dispatch/internal/domain/shared"
	"time"
)

type Assignment struct {
	ID      string
	OrderID string
	DroneID string

	Status AssignmentStatus
	Target *shared.Location

	CreatedAt time.Time
	UpdatedAt time.Time
}
