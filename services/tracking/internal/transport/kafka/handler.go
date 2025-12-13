package kafka

import (
	"context"
	"encoding/json"
	"hive/services/tracking/internal/models"
	"hive/services/tracking/internal/repository"
)

type Handler struct {
	repo *repository.TrackingRepository
}

func New(repo *repository.TrackingRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleMessage(ctx context.Context, data []byte) error {
	var drone models.Drone
	if err := json.Unmarshal(data, &drone); err != nil {
		return err
	}

	if err := h.repo.UpdateAllDrones(
		ctx, drone,
	); err != nil {
		return err
	}

	if err := h.repo.UpdateData(
		ctx, drone,
	); err != nil {
		return err
	}

	if err := h.repo.UpdateGeopostion(
		ctx, drone,
	); err != nil {
		return err
	}

	return nil
}
