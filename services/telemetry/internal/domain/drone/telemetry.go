package drone

import "hive/services/telemetry/internal/domain/shared"

type TelemetryData struct {
	DroneID       string          `json:"drone_id"`
	DroneLocation shared.Location `json:"drone_location"`
	Battery       float64         `json:"battery"`
	Status        DroneStatus     `json:"status"`
	Timestamp     int64           `json:"timestamp"`
	Event         DroneEvent      `json:"event"`
}

type TelemetryEvent struct {
	DroneID       string          `json:"drone_id"`
	Event         DroneEvent      `json:"event"`
	DroneLocation shared.Location `json:"drone_location"`
	Timestamp     int64           `json:"timestamp"`
}
