package server

import (
	"context"
	"fmt"
	"hive/gen/order"
	"hive/gen/tracking"
	"hive/services/api-gateway/internal/client"
	"hive/services/api-gateway/internal/config"
	"hive/services/api-gateway/internal/handler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo           *echo.Echo
	cfg            *config.Config
	orderClient    *order.OrderServiceClient
	trackingClient *tracking.TrackingServiceClient
	grpcClients    []interface{ Close() error }
}

func NewServer(cfg *config.Config) (*Server, error) {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	orderClient, err := client.NewOrderClient(cfg.OrderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create order client: %w", err)
	}

	trackingClient, err := client.NewTrackingClient(cfg.TrackingAddr)
	if err != nil {
		orderClient.Close()
		return nil, fmt.Errorf("failed to create tracking client: %w", err)
	}

	h := handler.NewHandler(&orderClient.client, &trackingClient.client)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")
	api.POST("/orders", h.CreateOrder)
	api.GET("/orders/:id", h.GetOrder)

	return &Server{
		echo:           e,
		cfg:            cfg,
		orderClient:    &orderClient.client,
		trackingClient: &trackingClient.client,
		grpcClients:    []interface{ Close() error }{orderClient, trackingClient},
	}, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.HTTPPort)
	fmt.Printf("Starting API Gateway on %s\n", addr)
	return s.echo.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	for _, client := range s.grpcClients {
		client.Close()
	}

	return s.echo.Shutdown(ctx)
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-quit:
		fmt.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		fmt.Println("Server stopped")
		return nil
	}
}
