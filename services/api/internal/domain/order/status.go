package order

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusAssigned  OrderStatus = "ASSIGNED"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusFailed    OrderStatus = "FAILED"
)
