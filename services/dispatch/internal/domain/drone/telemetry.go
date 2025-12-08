package drone

import "hive/services/dispatch/internal/domain/shared"

type TelemetryEvent struct {
	DroneID       string          `json:"drone_id"`
	Event         DroneEvent      `json:"event"`
	DroneLocation shared.Location `json:"drone_location"`
	Timestamp     int64           `json:"timestamp"`
}
