package drone

import "hive/services/dispatch/internal/domain/shared"

type DroneNearestInfo struct {
	ID       string
	Distance float64
}

type DroneInfo struct {
	ID                  string
	Location            shared.Location
	Battery             float64
	SpeedMps            float64
	ConsumptionPerMeter float64
}
