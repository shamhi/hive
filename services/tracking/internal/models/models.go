package models

type DroneStatus string

const (
	DroneStatusFree     DroneStatus = "FREE"
	DroneStatusBusy     DroneStatus = "BUSY"
	DroneStatusCharging DroneStatus = "CHARGING"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type TelemetryData struct {
	DroneID             string      `json:"drone_id"`
	DroneLocation       Location    `json:"drone_location"`
	Battery             float64     `json:"battery"`
	SpeedMps            float64     `json:"speed_mps"`
	ConsumptionPerMeter float64     `json:"consumption_per_meter"`
	Status              DroneStatus `json:"status"`
	Timestamp           int64       `json:"timestamp"`
}

type Drone struct {
	ID                  string      `json:"-"`
	Location            Location    `json:"location"`
	Battery             float64     `json:"battery"`
	Status              DroneStatus `json:"status"`
	SpeedMps            float64     `json:"speed_mps"`
	ConsumptionPerMeter float64     `json:"consumption_per_meter"`
}
