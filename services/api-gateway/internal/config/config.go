package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort     int
	OrderAddr    string
	TrackingAddr string
}

func Load() *Config {
	port := 8080
	if p := os.Getenv("HTTP_PORT"); p != "" {
		if pInt, err := strconv.Atoi(p); err == nil {
			port = pInt
		}
	}

	return &Config{
		HTTPPort:     port,
		OrderAddr:    getEnv("ORDER_ADDR", "localhost:50051"),
		TrackingAddr: getEnv("TRACKING_ADDR", "localhost:50052"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
