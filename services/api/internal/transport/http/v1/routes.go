package v1

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo, handler *Handler) {
	api := e.Group("/api/v1")

	e.GET("/ping", handler.Ping)

	api.POST("/orders", handler.CreateOrder)
	api.GET("/orders/:id", handler.GetOrder)
}
