package assignment

import (
	"hive/services/api-gateway/internal/domain/shared"
)

type Assignment struct {
	ID     string
	Status AssignmentStatus
	Target *shared.Location
}
