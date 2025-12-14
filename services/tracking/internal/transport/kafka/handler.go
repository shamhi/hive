package kafka

import (
	"context"
	"encoding/json"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/service"
)

type Handler struct {
	repo service.DroneRepository
}

func New(repo service.DroneRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleMessage(ctx context.Context, data []byte) error {
	var tm drone.TelemetryData
	if err := json.Unmarshal(data, &tm); err != nil {
		return err
	}

	if err := h.repo.UpdateState(ctx, tm); err != nil {
		return err
	}

	return nil
}
