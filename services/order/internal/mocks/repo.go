package mocks

import (
	"context"
	"hive/services/order/internal/domain/order"
	"hive/services/order/internal/service"
	"sync"
)

type MockRepo struct {
	mu     sync.RWMutex
	orders map[string]*order.Order
}

func NewMockRepo() *MockRepo {
	return &MockRepo{orders: map[string]*order.Order{}}
}

func (m *MockRepo) Save(_ context.Context, o *order.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders[o.ID] = o
	return nil
}

func (m *MockRepo) GetByID(_ context.Context, id string) (*order.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if o, exist := m.orders[id]; exist {
		return o, nil
	}
	return nil, service.ErrNotFound
}

func (m *MockRepo) UpdateStatus(_ context.Context, id string, status order.OrderStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	o, exist := m.orders[id]
	if !exist {
		return service.ErrNotFound
	}
	o.Status = status

	return nil
}

func (m *MockRepo) SetDroneID(_ context.Context, id string, droneID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	o, exist := m.orders[id]
	if !exist {
		return service.ErrNotFound
	}
	o.DroneID = droneID

	return nil
}

func (m *MockRepo) UpdateDroneAndStatus(ctx context.Context, id string, droneID string, status order.OrderStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	o, exist := m.orders[id]
	if !exist {
		return service.ErrNotFound
	}
	o.DroneID = droneID
	o.Status = status

	return nil
}
