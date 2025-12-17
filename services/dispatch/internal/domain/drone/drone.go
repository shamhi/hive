package drone

import "hive/services/dispatch/internal/domain/shared"

type Drone struct {
	ID                  string
	Battery             float64
	SpeedMps            float64
	ConsumptionPerMeter float64
	Status              DroneStatus
	Location            shared.Location
}

type DroneNearest struct {
	ID       string
	Distance float64
}
