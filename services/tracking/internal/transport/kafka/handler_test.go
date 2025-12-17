package kafka

import (
	"context"
	"encoding/json"
	"hive/pkg/logger"
	"testing"

	"hive/services/tracking/internal/domain/drone"
	"hive/services/tracking/internal/domain/shared"
	"hive/services/tracking/internal/service"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type repoStub struct {
	UpdateCalls int
	Last        drone.TelemetryData
	Err         error
}

func (r *repoStub) GetNearest(context.Context, shared.Location, float64, float64) (*drone.DroneNearest, error) {
	return nil, service.ErrDroneNotFound
}
func (r *repoStub) GetByID(context.Context, string) (*drone.Drone, error) {
	return nil, service.ErrDroneNotFound
}
func (r *repoStub) List(context.Context, int64, int64) ([]*drone.Drone, error) {
	return []*drone.Drone{}, nil
}
func (r *repoStub) SetStatus(context.Context, string, drone.DroneStatus) error { return nil }

func (r *repoStub) UpdateState(ctx context.Context, tm drone.TelemetryData) error {
	r.UpdateCalls++
	r.Last = tm
	return r.Err
}

func TestHandler_HandleMessage_OK(t *testing.T) {
	repo := &repoStub{}
	h := New(repo, lg)

	payload, _ := json.Marshal(drone.TelemetryData{
		DroneID:       "d1",
		DroneLocation: shared.Location{Lat: 55, Lon: 37},
		Battery:       50,
	})

	err := h.HandleMessage(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.UpdateCalls != 1 {
		t.Fatalf("expected 1 call")
	}
	if repo.Last.DroneID != "d1" {
		t.Fatalf("unexpected data")
	}
}

func TestHandler_HandleMessage_BadJSON(t *testing.T) {
	repo := &repoStub{}
	h := New(repo, lg)

	err := h.HandleMessage(context.Background(), []byte("{"))
	if err == nil {
		t.Fatalf("expected error")
	}
}
