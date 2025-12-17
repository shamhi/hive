package drone

import "hive/services/tracking/internal/domain/shared"

type Drone struct {
	ID                  string
	Battery             float64
	SpeedMps            float64
	ConsumptionPerMeter float64
	Status              DroneStatus
	Location            shared.Location
	UpdatedAt           int64
}

type DroneNearest struct {
	ID       string
	Distance float64
}
