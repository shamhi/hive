package domain

import "time"

type Location struct {
	Lat float64
	Lon float64
}

type Order struct {
	ID        string
	UserID    string
	DroneID   string
	Items     []string
	Status    OrderStatus
	Location  Location
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderStatus string

const (
	CREATED   OrderStatus = "CREATED"
	PENDING   OrderStatus = "PENDING"
	ASSIGNED  OrderStatus = "ASSIGNED"
	COMPLETED OrderStatus = "COMPLETED"
	FAILED    OrderStatus = "FAILED"
)
