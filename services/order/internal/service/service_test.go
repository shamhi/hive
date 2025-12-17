package service

import (
	"context"
	"errors"
	"hive/pkg/logger"
	"testing"

	"hive/services/order/internal/domain/assignment"
	"hive/services/order/internal/domain/order"
	"hive/services/order/internal/domain/shared"
)

var (
	lg, _ = logger.NewLogger("dev")
)

type orderRepoStub struct {
	SaveFn                 func(context.Context, *order.Order) error
	GetByIDFn              func(context.Context, string) (*order.Order, error)
	UpdateStatusFn         func(context.Context, string, order.OrderStatus) error
	SetDroneIDFn           func(context.Context, string, string) error
	UpdateDroneAndStatusFn func(context.Context, string, string, order.OrderStatus) error

	Saved                     *order.Order
	UpdateDroneAndStatusCalls int
	LastUpdateDroneID         string
	LastUpdateStatus          order.OrderStatus
}

func (r *orderRepoStub) Save(ctx context.Context, o *order.Order) error {
	r.Saved = o
	if r.SaveFn != nil {
		return r.SaveFn(ctx, o)
	}
	return nil
}

func (r *orderRepoStub) GetByID(ctx context.Context, id string) (*order.Order, error) {
	if r.GetByIDFn != nil {
		return r.GetByIDFn(ctx, id)
	}
	return nil, ErrOrderNotFound
}

func (r *orderRepoStub) UpdateStatus(ctx context.Context, id string, st order.OrderStatus) error {
	r.LastUpdateStatus = st
	if r.UpdateStatusFn != nil {
		return r.UpdateStatusFn(ctx, id, st)
	}
	return nil
}

func (r *orderRepoStub) SetDroneID(ctx context.Context, id string, droneID string) error {
	if r.SetDroneIDFn != nil {
		return r.SetDroneIDFn(ctx, id, droneID)
	}
	return nil
}

func (r *orderRepoStub) UpdateDroneAndStatus(ctx context.Context, id, droneID string, st order.OrderStatus) error {
	r.UpdateDroneAndStatusCalls++
	r.LastUpdateDroneID = droneID
	r.LastUpdateStatus = st
	if r.UpdateDroneAndStatusFn != nil {
		return r.UpdateDroneAndStatusFn(ctx, id, droneID, st)
	}
	return nil
}

type dispatchClientStub struct {
	AssignFn func(context.Context, string, *shared.Location) (*assignment.AssignmentInfo, error)
}

func (d *dispatchClientStub) AssignDrone(ctx context.Context, orderID string, loc *shared.Location) (*assignment.AssignmentInfo, error) {
	if d.AssignFn != nil {
		return d.AssignFn(ctx, orderID, loc)
	}
	return nil, errors.New("dispatch error")
}

func TestOrderService_CreateOrder_Validation(t *testing.T) {
	repo := &orderRepoStub{}
	dispatch := &dispatchClientStub{}
	svc := NewOrderService(repo, dispatch, lg)

	_, err := svc.CreateOrder(context.Background(), "", []string{"a"}, shared.Location{Lat: 55, Lon: 37})
	if err == nil {
		t.Fatalf("expected error")
	}

	_, err = svc.CreateOrder(context.Background(), "u", nil, shared.Location{Lat: 55, Lon: 37})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOrderService_CreateOrder_DispatchFails_FallbackPending(t *testing.T) {
	repo := &orderRepoStub{}
	dispatch := &dispatchClientStub{
		AssignFn: func(ctx context.Context, orderID string, loc *shared.Location) (*assignment.AssignmentInfo, error) {
			return nil, errors.New("down")
		},
	}
	svc := NewOrderService(repo, dispatch, lg)

	info, err := svc.CreateOrder(context.Background(), "u", []string{"a"}, shared.Location{Lat: 55, Lon: 37})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != order.OrderStatusPending {
		t.Fatalf("expected pending")
	}
	if info.DroneID != "" {
		t.Fatalf("expected empty drone id")
	}
	if repo.Saved == nil || repo.Saved.UserID != "u" {
		t.Fatalf("expected saved order")
	}
	if repo.UpdateDroneAndStatusCalls != 0 {
		t.Fatalf("expected no UpdateDroneAndStatus call")
	}
}

func TestOrderService_CreateOrder_Assigned(t *testing.T) {
	repo := &orderRepoStub{}
	dispatch := &dispatchClientStub{
		AssignFn: func(ctx context.Context, orderID string, loc *shared.Location) (*assignment.AssignmentInfo, error) {
			return &assignment.AssignmentInfo{DroneID: "d1", EtaSeconds: 10}, nil
		},
	}
	svc := NewOrderService(repo, dispatch, lg)

	info, err := svc.CreateOrder(context.Background(), "u", []string{"a"}, shared.Location{Lat: 55, Lon: 37})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != order.OrderStatusAssigned {
		t.Fatalf("expected assigned")
	}
	if info.DroneID != "d1" {
		t.Fatalf("expected drone d1")
	}
	if repo.UpdateDroneAndStatusCalls != 1 {
		t.Fatalf("expected UpdateDroneAndStatus called")
	}
	if repo.LastUpdateDroneID != "d1" || repo.LastUpdateStatus != order.OrderStatusAssigned {
		t.Fatalf("unexpected update values")
	}
}

func TestOrderService_GetOrder_NotFoundWrapped(t *testing.T) {
	repo := &orderRepoStub{
		GetByIDFn: func(ctx context.Context, id string) (*order.Order, error) {
			return nil, ErrOrderNotFound
		},
	}
	svc := NewOrderService(repo, &dispatchClientStub{}, lg)

	_, err := svc.GetOrder(context.Background(), "x")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound")
	}
}
