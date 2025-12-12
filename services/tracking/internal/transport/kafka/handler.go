package kafka

import (
	"context"
	"encoding/json"
	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/repository"
)

type Handler struct {
	repo *repository.TrackingRepository
}

func New(repo *repository.TrackingRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleMessage(ctx context.Context, data []byte) error {
	// TODO: передавать структуру drone.TelemetryData и работать с ней
	var msg drone.TelemetryData
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	if err := h.repo.UpdateGeo(
		ctx,
		msg.DroneID,
		msg.DroneLocation.Lon,
		msg.DroneLocation.Lat,
	); err != nil {
		return err
	}

	if err := h.repo.UpdateState(
		ctx,
		msg.DroneID,
		msg.Battery,
		string(msg.Status),
	); err != nil {
		return err
	}

	return nil
}
