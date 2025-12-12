package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hive/services/api/internal/clients"
	"hive/services/api/internal/config"
	"hive/services/api/internal/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/go-playground/validator.v9"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	orderClient, err := clients.NewOrderClient(cfg.OrderService.Address, cfg.OrderService.Timeout)
	if err != nil {
		log.Fatalf("Failed to create order client: %v", err)
	}
	defer func() {
		if err := orderClient.Close(); err != nil {
			log.Printf("Error closing order client: %v", err)
		}
	}()

	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | ${method} | ${uri} | ${status} | ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	handler := handlers.NewHandler(orderClient)
	handlers.RegisterRoutes(e, handler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := ":" + string(rune(cfg.Server.Port))
		log.Printf("API Gateway starting on port %d", cfg.Server.Port)
		log.Printf("Order Service address: %s", cfg.OrderService.Address)
		if err := e.Start(addr); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("API Gateway stopped gracefully")
}
