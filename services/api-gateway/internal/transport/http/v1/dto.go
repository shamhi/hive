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
	DroneID             string         `json:"drone_id"`
	Battery             float64        `json:"battery"`
	SpeedMps            float64        `json:"speed_mps"`
	ConsumptionPerMeter float64        `json:"consumption_per_meter"`
	Status              string         `json:"status"`
	UpdatedAtMs         int64          `json:"updated_at_ms"`
	Location            Location       `json:"location"`
	Assignment          *AssignmentDTO `json:"assignment,omitempty"`
}

type AssignmentDTO struct {
	AssignmentID   string    `json:"assignment_id"`
	Status         string    `json:"status"`
	TargetLocation *Location `json:"target_location"`
}

type ListBasesResponse struct {
	Items []BaseDTO `json:"items"`
}

type ListStoresResponse struct {
	Items []StoreDTO `json:"items"`
}

type ListDroneResponse struct {
	ServerTimeMs int64      `json:"server_time_ms"`
	Items        []DroneDTO `json:"items"`
}
