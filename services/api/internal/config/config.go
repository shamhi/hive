package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server       ServerConfig
	OrderService OrderServiceConfig
}

type ServerConfig struct {
	Port int
}

type OrderServiceConfig struct {
	Address string
	Timeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		OrderService: OrderServiceConfig{
			Address: getEnv("ORDER_SERVICE_ADDR", "order:50051"),
			Timeout: getEnvAsDuration("ORDER_SERVICE_TIMEOUT", 30*time.Second),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		panic(fmt.Sprintf("Invalid value for %s: %s", key, strValue))
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(strValue)
	if err != nil {
		panic(fmt.Sprintf("Invalid duration for %s: %s", key, strValue))
	}

	return value
}
