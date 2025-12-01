package mocks

import (
	"context"
	"hive/services/order/internal/domain"
	"hive/services/order/internal/service"
	"sync"
)

type MockRepo struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

func NewMockRepo() *MockRepo {
	return &MockRepo{orders: map[string]*domain.Order{}}
}

func (m *MockRepo) Save(_ context.Context, o *domain.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders[o.ID] = o
	return nil
}

func (m *MockRepo) Get(_ context.Context, id string) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if o, exist := m.orders[id]; exist {
		return o, nil
	}
	return nil, service.ErrNotFound
}

func (m *MockRepo) Update(_ context.Context, o *domain.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stgO, exist := m.orders[o.ID]
	if !exist {
		return service.ErrNotFound
	}
	stgO.DroneID = o.DroneID
	stgO.Items = o.Items
	stgO.Status = o.Status
	stgO.Location = o.Location

	return nil
}
