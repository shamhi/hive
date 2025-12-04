package handler

import (
	"fmt"
	"hive/gen/order"
	"hive/gen/tracking"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type Handler struct {
	client order.OrderServiceClient
	conn   *grpc.ClientConn
	// orderClient    *order.OrderServiceClient
	// trackingClient *tracking.TrackingServiceClient
}

func NewHandler(orderClient *order.OrderServiceClient, trackingClient *tracking.TrackingServiceClient) *Handler {
	return &Handler{
		orderClient:    orderClient,
		trackingClient: trackingClient,
	}
}

func (h *Handler) CreateOrder(c echo.Context) error {
	var req struct {
		UserID           string   `json:"user_id"`
		Items            []string `json:"items"`
		DeliveryLocation struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"delivery_location"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.UserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}
	if len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "items cannot be empty",
		})
	}

	orderReq := &order.CreateOrderRequest{
		UserId: req.UserID,
		Items:  req.Items,
		DeliveryLocation: &order.Location{
			Lat: req.DeliveryLocation.Lat,
			Lon: req.DeliveryLocation.Lon,
		},
	}

	resp, err := (*h.orderClient).CreateOrder(c.Request().Context(), orderReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to create order: %v", err),
		})
	}

	response := map[string]interface{}{
		"order_id":       resp.GetOrderId(),
		"status":         resp.GetStatus().String(),
		"estimated_time": "15 min",
	}

	if resp.GetDroneId() != "" {
		response["drone_id"] = resp.GetDroneId()
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetOrder(c echo.Context) error {
	orderID := c.Param("id")
	if strings.TrimSpace(orderID) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "order ID is required",
		})
	}

	orderReq := &order.GetOrderRequest{OrderId: orderID}
	orderResp, err := (*h.orderClient).GetOrder(c.Request().Context(), orderReq)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "order not found",
		})
	}

	response := map[string]interface{}{
		"order_id": orderResp.GetOrderId(),
		"status":   orderResp.GetStatus().String(),
	}

	if droneID := orderResp.GetDroneId(); droneID != "" {
		trackingReq := &tracking.GetDroneLocationRequest{DroneId: droneID}
		trackingResp, err := (*h.trackingClient).GetDroneLocation(c.Request().Context(), trackingReq)

		droneInfo := map[string]interface{}{
			"id": droneID,
		}

		if err == nil && trackingResp != nil {
			if loc := trackingResp.GetDroneLocation(); loc != nil {
				droneInfo["location"] = map[string]float64{
					"lat": loc.GetLat(),
					"lon": loc.GetLon(),
				}
			}
			droneInfo["battery"] = trackingResp.GetBattery()
		}

		response["drone"] = droneInfo
	}

	return c.JSON(http.StatusOK, response)
}
