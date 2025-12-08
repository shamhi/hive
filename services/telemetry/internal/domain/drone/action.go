package drone

type DroneAction string

const (
	DroneActionWait        DroneAction = "WAIT"
	DroneActionFlyTo       DroneAction = "FLY_TO"
	DroneActionPickupCargo DroneAction = "PICKUP_CARGO"
	DroneActionDropCargo   DroneAction = "DROP_CARGO"
	DroneActionCharge      DroneAction = "CHARGE"
)
