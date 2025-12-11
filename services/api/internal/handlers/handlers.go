package handlers

import (
	"hive/services/api/internal/models"
	"hive/services/api/repository"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type OrderHandler struct {
	repo *repository.OrderRepository
}

func NewOrderHandler(repo *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{repo: repo}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
	var req models.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user_id format"})
	}

	orderID := uuid.New()
	now := time.Now()

	order := &models.Order{
		ID:          orderID,
		UserID:      userID,
		Items:       models.StringSlice(req.Items),
		DeliveryLat: req.DeliveryLocation.Lat,
		DeliveryLon: req.DeliveryLocation.Lon,
		Status:      "PENDING",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if time.Now().UnixNano()%10 < 7 {
		order.Status = "ASSIGNED"
		droneID := uuid.New()
		order.DroneID = &droneID
		estimatedTime := "15min"
		order.EstimatedTime = &estimatedTime
	}

	ctx := c.Request().Context()
	if err := h.repo.CreateOrder(ctx, order); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order"})
	}
	resp := models.CreateOrderResponse{
		OrderID: order.ID.String(),
		Status:  order.Status,
	}
	if order.EstimatedTime != nil {
		resp.EstimatedTime = *order.EstimatedTime
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *OrderHandler) GetOrder(c echo.Context) error {
	orderIDStr := c.Param("id")

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID format"})
	}

	ctx := c.Request().Context()
	order, err := h.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get order"})
	}

	resp := models.GetOrderResponse{
		OrderID: order.ID.String(),
		Status:  order.Status,
	}

	if order.DroneID != nil {
		resp.Drone = &models.Drone{
			ID: *order.DroneID,
			Location: models.Location{
				Lat: 55.7289473,
				Lon: 37.7457302,
			},
			Battery: 85,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) GetUserOrders(c echo.Context) error {
	userIDStr := c.QueryParam("user_id")
	if userIDStr == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user_id is required"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user_id format"})
	}

	ctx := c.Request().Context()
	orders, err := h.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user orders"})
	}

	return c.JSON(http.StatusOK, orders)
}
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":    "ok",
		"service":   "drone-delivery-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func RegisterRoutes(e *echo.Echo, repo *repository.OrderRepository) {
	handler := NewOrderHandler(repo)

	api := e.Group("/api/v1")

	e.GET("/health", HealthCheck)

	ordersGroup := api.Group("/orders")
	ordersGroup.POST("", handler.CreateOrder)
	ordersGroup.GET("/:id", handler.GetOrder)
	ordersGroup.GET("", handler.GetUserOrders)

	ordersGroup.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))
}
