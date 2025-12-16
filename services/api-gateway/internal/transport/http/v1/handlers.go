package v1

import (
	"hive/services/api-gateway/internal/domain/shared"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	order    OrderClient
	base     BaseClient
	store    StoreClient
	tracking TrackingClient
}

func NewHandler(
	order OrderClient,
	base BaseClient,
	store StoreClient,
	tracking TrackingClient,
) *Handler {
	return &Handler{
		order:    order,
		base:     base,
		store:    store,
		tracking: tracking,
	}
}

func (h *Handler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
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
	orderInfo, err := h.order.CreateOrder(
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
	o, err := h.order.GetOrder(ctx, orderID)
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

func (h *Handler) ListBases(c echo.Context) error {
	// Implementation goes here
	return nil
}

func (h *Handler) ListStores(c echo.Context) error {
	// Implementation goes here
	return nil
}

func (h *Handler) ListDrones(c echo.Context) error {
	// Implementation goes here
	return nil
}
