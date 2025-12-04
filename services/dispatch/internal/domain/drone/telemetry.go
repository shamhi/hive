package drone

type TelemetryEvent struct {
	DroneID   string     `json:"drone_id"`
	Event     DroneEvent `json:"event"`
	Timestamp int64      `json:"timestamp"`
}
