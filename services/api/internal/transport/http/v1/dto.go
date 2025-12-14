package v1

type Location struct {
	Lat float64 `json:"lat" validate:"required,latitude"`
	Lon float64 `json:"lon" validate:"required,longitude"`
}

type CreateOrderRequest struct {
	UserID           string   `json:"user_id" validate:"required,uuid4"`
	Items            []string `json:"items" validate:"required,min=1"`
	DeliveryLocation Location `json:"delivery_location" validate:"required"`
}

type CreateOrderResponse struct {
	OrderID    string `json:"order_id"`
	Status     string `json:"status"`
	DroneID    string `json:"drone_id,omitempty"`
	EtaSeconds int32  `json:"eta_seconds,omitempty"`
}

type GetOrderResponse struct {
	OrderID  string   `json:"order_id"`
	UserID   string   `json:"user_id"`
	DroneID  string   `json:"drone_id,omitempty"`
	Items    []string `json:"items"`
	Status   string   `json:"status"`
	Location Location `json:"delivery_location"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
