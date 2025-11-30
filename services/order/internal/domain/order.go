package domain

type Location struct {
	Lat float64
	Lon float64
}

type Order struct {
	ID       string
	Items    []string
	Status   OrderStatus
	Location Location
}

type OrderStatus string

const (
	CREATED   OrderStatus = "CREATED"
	PENDING   OrderStatus = "PENDING"
	ASSIGNED  OrderStatus = "ASSIGNED"
	COMPLETED OrderStatus = "COMPLETED"
	FAILED    OrderStatus = "FAILED"
)
