package drone

import "hive/services/tracking/internal/domain/shared"

type TelemetryData struct {
	DroneID             string          `json:"drone_id"`
	DroneLocation       shared.Location `json:"drone_location"`
	Battery             float64         `json:"battery"`
	SpeedMps            float64         `json:"speed_mps"`
	ConsumptionPerMeter float64         `json:"consumption_per_meter"`
	Status              DroneStatus     `json:"status"`
}
