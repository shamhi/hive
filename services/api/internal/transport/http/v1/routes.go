package v1

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo, handler *Handler) {
	api := e.Group("/api/v1")

	e.GET("/health", handler.HealthCheck)

	api.POST("/orders", handler.CreateOrder)
	api.GET("/orders/:id", handler.GetOrder)
}
