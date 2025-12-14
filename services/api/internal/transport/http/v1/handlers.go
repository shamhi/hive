package v1

import (
	"hive/services/api/internal/domain/shared"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	orderClient OrderClient
}

func NewHandler(orderClient OrderClient) *Handler {
	return &Handler{
		orderClient: orderClient,
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
	orderInfo, err := h.orderClient.CreateOrder(
		ctx,
		req.UserID,
		req.Items,
		shared.Location{
			Lat: req.DeliveryLocation.Lat,
			Lon: req.DeliveryLocation.Lon,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order: " + err.Error()})
	}

	resp := CreateOrderResponse{
		OrderID:    orderInfo.ID,
		Status:     string(orderInfo.Status),
		DroneID:    orderInfo.DroneID,
		EtaSeconds: orderInfo.EtaSeconds,
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetOrder(c echo.Context) error {
	orderID := c.Param("id")

	if _, err := uuid.Parse(orderID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order_id format"})
	}

	ctx := c.Request().Context()
	o, err := h.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get order: " + err.Error()})
	}

	resp := GetOrderResponse{
		OrderID: o.ID,
		UserID:  o.UserID,
		DroneID: o.DroneID,
		Items:   o.Items,
		Status:  string(o.Status),
		Location: Location{
			Lat: o.Location.Lat,
			Lon: o.Location.Lon,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
