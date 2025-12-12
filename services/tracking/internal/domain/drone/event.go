package drone

type DroneEvent string

const (
	DroneEventNone            DroneEvent = "EVENT_NONE"
	DroneEventArrivedAtStore  DroneEvent = "EVENT_ARRIVED_AT_STORE"
	DroneEventPickedUpCargo   DroneEvent = "EVENT_PICKED_UP_CARGO"
	DroneEventArrivedAtClient DroneEvent = "EVENT_ARRIVED_AT_CLIENT"
	DroneEventDroppedCargo    DroneEvent = "EVENT_DROPPED_CARGO"
	DroneEventArrivedAtBase   DroneEvent = "EVENT_ARRIVED_AT_BASE"
	DroneEventFullyCharged    DroneEvent = "EVENT_FULLY_CHARGED"
)
