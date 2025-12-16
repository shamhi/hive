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

type BaseDTO struct {
	BaseID   string   `json:"base_id"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}
type StoreDTO struct {
	StoreID  string   `json:"store_id"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}
type DroneDTO struct {
	DroneID             string   `json:"drone_id"`
	Location            Location `json:"location"`
	Battery             float64  `json:"battery"`
	SpeedMps            float64  `json:"speed_mps"`
	ConsumptionPerMeter float64  `json:"consumption_per_meter"`
}
type ListBasesResponse struct {
	Bases []BaseDTO `json:"bases"`
}
type ListStoresResponse struct {
	Stores []StoreDTO `json:"stores"`
}

type ListDroneResponse struct {
	Drones []DroneDTO `json:"drones"`
}
