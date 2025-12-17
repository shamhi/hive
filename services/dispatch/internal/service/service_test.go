package service

import (
	"context"
	"errors"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"testing"
	"time"

	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/domain/base"
	"hive/services/dispatch/internal/domain/drone"
	"hive/services/dispatch/internal/domain/order"
	"hive/services/dispatch/internal/domain/shared"
	"hive/services/dispatch/internal/domain/store"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type repoStub struct {
	SaveFn         func(context.Context, *assignment.Assignment) error
	UpdateStatusFn func(context.Context, string, assignment.AssignmentStatus) error

	Saved    *assignment.Assignment
	Statuses []assignment.AssignmentStatus
}

func (r *repoStub) Save(ctx context.Context, a *assignment.Assignment) error {
	r.Saved = a
	if r.SaveFn != nil {
		return r.SaveFn(ctx, a)
	}
	return nil
}

func (r *repoStub) GetByID(context.Context, string) (*assignment.Assignment, error) {
	return nil, errors.New("not used")
}
func (r *repoStub) GetByDroneID(context.Context, string) (*assignment.Assignment, error) {
	return nil, errors.New("not used")
}

func (r *repoStub) UpdateStatus(ctx context.Context, id string, st assignment.AssignmentStatus) error {
	r.Statuses = append(r.Statuses, st)
	if r.UpdateStatusFn != nil {
		return r.UpdateStatusFn(ctx, id, st)
	}
	return nil
}

type storeClientStub struct {
	FindNearestFn func(context.Context, *shared.Location) (*store.StoreNearest, error)
	GetLocFn      func(context.Context, string) (*store.Store, error)
}

func (s *storeClientStub) FindNearest(ctx context.Context, loc *shared.Location) (*store.StoreNearest, error) {
	return s.FindNearestFn(ctx, loc)
}
func (s *storeClientStub) GetStoreLocation(ctx context.Context, id string) (*store.Store, error) {
	return s.GetLocFn(ctx, id)
}

type baseClientStub struct {
	FindNearestFn func(context.Context, *shared.Location) (*base.BaseNearest, error)
	GetLocFn      func(context.Context, string) (*base.Base, error)
}

func (b *baseClientStub) FindNearest(ctx context.Context, loc *shared.Location) (*base.BaseNearest, error) {
	return b.FindNearestFn(ctx, loc)
}

func (b *baseClientStub) GetBaseLocation(ctx context.Context, id string) (*base.Base, error) {
	return b.GetLocFn(ctx, id)
}

type trackingClientStub struct {
	FindNearestFn func(context.Context, *shared.Location, float64, float64) (*drone.DroneNearest, error)
	GetLocFn      func(context.Context, string) (*drone.Drone, error)
	SetStatusFn   func(context.Context, string, drone.DroneStatus) error

	SetCalls []drone.DroneStatus
}

func (t *trackingClientStub) FindNearest(ctx context.Context, storeLoc *shared.Location, minBattery, radius float64) (*drone.DroneNearest, error) {
	return t.FindNearestFn(ctx, storeLoc, minBattery, radius)
}

func (t *trackingClientStub) GetDroneLocation(ctx context.Context, id string) (*drone.Drone, error) {
	return t.GetLocFn(ctx, id)
}

func (t *trackingClientStub) SetStatus(ctx context.Context, id string, st drone.DroneStatus) error {
	t.SetCalls = append(t.SetCalls, st)
	return t.SetStatusFn(ctx, id, st)
}

type telemetryClientStub struct {
	SendFn func(context.Context, string, drone.DroneAction, *drone.Target) error
	Calls  int
}

func (t *telemetryClientStub) SendCommand(ctx context.Context, id string, a drone.DroneAction, target *drone.Target) error {
	t.Calls++
	return t.SendFn(ctx, id, a, target)
}

type orderClientStub struct {
	UpdateFn func(context.Context, string, order.OrderStatus) error
	Calls    int
	Last     order.OrderStatus
}

func (o *orderClientStub) UpdateStatus(ctx context.Context, id string, st order.OrderStatus) error {
	o.Calls++
	o.Last = st
	return o.UpdateFn(ctx, id, st)
}

func TestDispatchService_AssignDrone_HappyPath(t *testing.T) {
	repo := &repoStub{}
	storeCl := &storeClientStub{
		FindNearestFn: func(ctx context.Context, loc *shared.Location) (*store.StoreNearest, error) {
			return &store.StoreNearest{ID: "s1", Distance: 200}, nil
		},
		GetLocFn: func(ctx context.Context, id string) (*store.Store, error) {
			return &store.Store{ID: id, Location: shared.Location{Lat: 55.7, Lon: 37.6}}, nil
		},
	}
	baseCl := &baseClientStub{
		FindNearestFn: func(ctx context.Context, loc *shared.Location) (*base.BaseNearest, error) {
			return &base.BaseNearest{ID: "b1", Distance: 50}, nil
		},
	}
	trackingCl := &trackingClientStub{
		FindNearestFn: func(ctx context.Context, loc *shared.Location, minB, rad float64) (*drone.DroneNearest, error) {
			return &drone.DroneNearest{ID: "d1", Distance: 100}, nil
		},
		GetLocFn: func(ctx context.Context, id string) (*drone.Drone, error) {
			return &drone.Drone{ID: id, Battery: 100, SpeedMps: 10, ConsumptionPerMeter: 0.1}, nil
		},
		SetStatusFn: func(ctx context.Context, id string, st drone.DroneStatus) error {
			return nil
		},
	}
	telemetryCl := &telemetryClientStub{
		SendFn: func(ctx context.Context, id string, a drone.DroneAction, target *drone.Target) error {
			return nil
		},
	}
	orderCl := &orderClientStub{
		UpdateFn: func(ctx context.Context, id string, st order.OrderStatus) error {
			return nil
		},
	}

	svc := NewDispatchService(
		repo,
		orderCl,
		storeCl,
		baseCl,
		trackingCl,
		telemetryCl,
		lg,
	)

	info, err := svc.AssignDrone(context.Background(), "o1", &shared.Location{Lat: 55.75, Lon: 37.61}, 10, 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.DroneID != "d1" {
		t.Fatalf("expected drone d1")
	}
	if info.EtaSeconds != int32(35) {
		t.Fatalf("expected eta=35, got %d", info.EtaSeconds)
	}
	if repo.Saved == nil || repo.Saved.OrderID != "o1" {
		t.Fatalf("expected saved assignment")
	}
	if len(repo.Statuses) < 2 || repo.Statuses[0] != assignment.AssignmentStatusAssigned || repo.Statuses[1] != assignment.AssignmentStatusFlyingToStore {
		t.Fatalf("expected statuses assigned, flying")
	}
	if telemetryCl.Calls != 1 {
		t.Fatalf("expected SendCommand called once")
	}
}

func TestDispatchService_AssignDrone_SetStatusFails_Rollback(t *testing.T) {
	repo := &repoStub{}
	storeCl := &storeClientStub{
		FindNearestFn: func(ctx context.Context, loc *shared.Location) (*store.StoreNearest, error) {
			return &store.StoreNearest{ID: "s1", Distance: 200}, nil
		},
		GetLocFn: func(ctx context.Context, id string) (*store.Store, error) {
			return &store.Store{ID: id, Location: shared.Location{Lat: 55.7, Lon: 37.6}}, nil
		},
	}
	baseCl := &baseClientStub{
		FindNearestFn: func(ctx context.Context, loc *shared.Location) (*base.BaseNearest, error) {
			return &base.BaseNearest{ID: "b1", Distance: 50}, nil
		},
	}
	trackingCl := &trackingClientStub{
		FindNearestFn: func(ctx context.Context, loc *shared.Location, minB, rad float64) (*drone.DroneNearest, error) {
			return &drone.DroneNearest{ID: "d1", Distance: 100}, nil
		},
		GetLocFn: func(ctx context.Context, id string) (*drone.Drone, error) {
			return &drone.Drone{ID: id, Battery: 100, SpeedMps: 10, ConsumptionPerMeter: 0.1}, nil
		},
		SetStatusFn: func(ctx context.Context, id string, st drone.DroneStatus) error {
			if st == drone.DroneStatusBusy {
				return errors.New("fail")
			}
			return nil
		},
	}
	telemetryCl := &telemetryClientStub{SendFn: func(context.Context, string, drone.DroneAction, *drone.Target) error { return nil }}
	orderCl := &orderClientStub{UpdateFn: func(context.Context, string, order.OrderStatus) error { return nil }}

	svc := NewDispatchService(
		repo,
		orderCl,
		storeCl,
		baseCl,
		trackingCl,
		telemetryCl,
		lg,
	)

	_, err := svc.AssignDrone(context.Background(), "o1", &shared.Location{Lat: 55.75, Lon: 37.61}, 10, 5000)
	if err == nil {
		t.Fatalf("expected error")
	}
	if len(trackingCl.SetCalls) < 2 {
		t.Fatalf("expected busy and free attempts")
	}
	foundFailed := false
	for _, st := range repo.Statuses {
		if st == assignment.AssignmentStatusFailed {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Fatalf("expected assignment failed status")
	}
}

func TestDispatchService_HandleTelemetryEvent_FullyCharged_RetryConfigFast(t *testing.T) {
	old := retryCfg
	retryCfg = resilience.RetryConfig{MaxAttempts: 1, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond, Jitter: 0}
	defer func() { retryCfg = old }()

	trackingCl := &trackingClientStub{
		SetStatusFn: func(ctx context.Context, id string, st drone.DroneStatus) error {
			return errors.New("down")
		},
	}
	svc := NewDispatchService(
		&repoStub{},
		&orderClientStub{},
		&storeClientStub{},
		&baseClientStub{},
		trackingCl,
		&telemetryClientStub{},
		lg,
	)

	err := svc.HandleTelemetryEvent(context.Background(), drone.TelemetryEvent{DroneID: "d1", Event: drone.DroneEventFullyCharged})
	if err == nil {
		t.Fatalf("expected error")
	}
}
