package kafka

import (
	"encoding/json"
	"fmt"
	"hive/services/dispatch/internal/service"
)

type TelemetryEvent struct {
	DroneID   string `json:"drone_id"`
	Event     string `json:"event"`
	Timestamp int64  `json:"timestamp"`
}

type Handler struct {
	dispatch *service.DispatchService
}

func NewHandler(dispatch *service.DispatchService) *Handler {
	return &Handler{
		dispatch: dispatch,
	}
}

func (h *Handler) Handle(msg []byte) error {
	var event TelemetryEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to unmarshal telemetry event: %w", err)
	}

	return h.dispatch.Handl
}
