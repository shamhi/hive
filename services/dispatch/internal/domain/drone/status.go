package drone

type DroneStatus string

const (
	DroneStatusFree     DroneStatus = "FREE"
	DroneStatusBusy     DroneStatus = "BUSY"
	DroneStatusCharging DroneStatus = "CHARGING"
)
