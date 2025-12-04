package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/service"
)

type Handler struct {
	dispatch *service.DispatchService
}

func NewHandler(dispatch *service.DispatchService) *Handler {
	return &Handler{
		dispatch: dispatch,
	}
}

func (h *Handler) Handle(ctx context.Context, msg []byte) error {
	var event drone.TelemetryEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("unmarshal telemetry: %w", err)
	}

	return h.dispatch.HandleTelemetryEvent(ctx, event)
}
