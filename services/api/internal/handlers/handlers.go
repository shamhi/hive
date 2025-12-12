package handlers

import (
	"net/http"
	"time"

	"hive/services/api/internal/clients"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	orderClient *clients.OrderClient
}

func NewHandler(orderClient *clients.OrderClient) *Handler {
	return &Handler{
		orderClient: orderClient,
	}
}

type CreateOrderRequest struct {
	UserID           string   `json:"user_id" validate:"required,uuid4"`
	Items            []string `json:"items" validate:"required,min=1"`
	DeliveryLocation Location `json:"delivery_location" validate:"required"`
}

type Location struct {
	Lat float64 `json:"lat" validate:"required,latitude"`
	Lon float64 `json:"lon" validate:"required,longitude"`
}

type CreateOrderResponse struct {
	OrderID    string `json:"order_id"`
	Status     string `json:"status"`
	DroneID    string `json:"drone_id,omitempty"`
	EtaSeconds int32  `json:"eta_seconds,omitempty"`
}

type GetOrderResponse struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	DroneID   string    `json:"drone_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func orderStatusToString(status int32) string {
	switch status {
	case 0:
		return "UNKNOWN"
	case 1:
		return "CREATED"
	case 2:
		return "PENDING"
	case 3:
		return "ASSIGNED"
	case 4:
		return "COMPLETED"
	case 5:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

func (h *Handler) CreateOrder(c echo.Context) error {
	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user_id format"})
	}

	ctx := c.Request().Context()
	resp, err := h.orderClient.CreateOrder(ctx, req.UserID, req.Items, req.DeliveryLocation.Lat, req.DeliveryLocation.Lon)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order: " + err.Error()})
	}

	apiResp := CreateOrderResponse{
		OrderID:    resp.GetOrderId(),
		Status:     orderStatusToString(int32(resp.GetStatus())),
		DroneID:    resp.GetDroneId(),
		EtaSeconds: resp.GetEtaSeconds(),
	}

	return c.JSON(http.StatusCreated, apiResp)
}

func (h *Handler) GetOrder(c echo.Context) error {
	orderID := c.Param("id")

	if _, err := uuid.Parse(orderID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order_id format"})
	}

	ctx := c.Request().Context()
	resp, err := h.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get order: " + err.Error()})
	}

	createdAt := time.Unix(resp.GetCreatedAt(), 0)
	updatedAt := time.Unix(resp.GetUpdatedAt(), 0)

	apiResp := GetOrderResponse{
		OrderID:   resp.GetOrderId(),
		Status:    orderStatusToString(int32(resp.GetStatus())),
		DroneID:   resp.GetDroneId(),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return c.JSON(http.StatusOK, apiResp)
}

func (h *Handler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "api-gateway",
	})
}

func RegisterRoutes(e *echo.Echo, handler *Handler) {
	api := e.Group("/api/v1")

	e.GET("/health", handler.HealthCheck)

	api.POST("/orders", handler.CreateOrder)
	api.GET("/orders/:id", handler.GetOrder)
}
