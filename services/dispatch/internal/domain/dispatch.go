package domain

import "time"

type Assignment struct {
	ID        string
	OrderID   string
	DroneID   string
	Status    AssignmentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Location struct {
	Lat float64
	Lon float64
}

type AssignmentStatus string

const (
	AssignmentStatusCreated   AssignmentStatus = "CREATED"
	AssignmentStatusAssigned  AssignmentStatus = "ASSIGNED"
	AssignmentStatusCompleted AssignmentStatus = "COMPLETED"
	AssignmentStatusFailed    AssignmentStatus = "FAILED"
)

type DroneStatus string

const (
	DroneStatusFree     DroneStatus = "FREE"
	DroneStatusBusy     DroneStatus = "BUSY"
	DroneStatusCharging DroneStatus = "CHARGING"
)

type DroneAction string

const (
	DroneActionWait        DroneAction = "WAIT"
	DroneActionFlyTo       DroneAction = "FLY_TO"
	DroneActionPickupCargo DroneAction = "PICKUP_CARGO"
	DroneActionDropCargo   DroneAction = "DROP_CARGO"
	DroneActionCharge      DroneAction = "CHARGE"
)

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusAssigned  OrderStatus = "ASSIGNED"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusFailed    OrderStatus = "FAILED"
)
