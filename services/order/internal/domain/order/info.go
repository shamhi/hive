package order

type OrderInfo struct {
	ID         string
	DroneID    string
	Status     OrderStatus
	EtaSeconds int32
}
