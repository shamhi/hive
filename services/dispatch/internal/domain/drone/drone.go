package drone

import "hive/services/dispatch/internal/domain/shared"

type Drone struct {
	ID                  string
	Location            shared.Location
	Battery             float64
	SpeedMps            float64
	ConsumptionPerMeter float64
}

type DroneNearest struct {
	ID       string
	Distance float64
}
