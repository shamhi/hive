package base

import "hive/services/dispatch/internal/domain/shared"

type BaseNearestInfo struct {
	ID       string
	Distance float64
}

type BaseInfo struct {
	ID                  string
	Location            shared.Location
	Battery             float64
	SpeedMps            float64
	ConsumptionPerMeter float64
}
