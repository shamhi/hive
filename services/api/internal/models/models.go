package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID               uuid.UUID   `json:"order_id" db:"id"`
	UserID           uuid.UUID   `json:"user_id" db:"user_id"`
	Items            StringSlice `json:"items" db:"items"`
	DeliveryLat      float64     `json:"-" db:"delivery_lat"`
	DeliveryLon      float64     `json:"-" db:"delivery_long"`
	DeliveryLocation Location    `json:"delivery_location" db:"-"`
	Status           string      `json:"status" db:"status"`
	DroneID          *uuid.UUID  `json:"drone_id,omitempty" db:"drone_id"`
	EstimatedTime    *string     `json:"estimated_time,omitempty" db:"estimated_time"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" db:"updated_at"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Drone struct {
	ID       uuid.UUID `json:"drone_id"`
	Location Location  `json:"location"`
	Battery  int       `json:"battery"`
}

type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("invalid type for StringSlice")
	}
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}

	return json.Marshal(s)
}

func (o *Order) SetDeliveryLocation() {
	o.DeliveryLocation = Location{
		Lat: o.DeliveryLat,
		Lon: o.DeliveryLon,
	}
}

type CreateOrderRequest struct {
	UserID           string   `json:"user_id" validate:"required,uuid4"`
	Items            []string `json:"items" validate:"required,min=1"`
	DeliveryLocation Location `json:"delivery_location" validate:"required"`
}

type CreateOrderResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	EstimatedTime string `json:"estimated_time,omitempty"`
}

type GetOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Drone   *Drone `json:"drone,omitempty"`
}
