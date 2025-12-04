package mocks

import (
	"context"
	"hive/services/dispatch/internal/domain/assignment"
	"hive/services/dispatch/internal/service"
	"sync"
)

type MockAssignmentRepo struct {
	mu    sync.RWMutex
	store map[string]*assignment.Assignment
}

func NewMockAssignmentRepo() *MockAssignmentRepo {
	return &MockAssignmentRepo{
		store: make(map[string]*assignment.Assignment),
	}
}

func (m *MockAssignmentRepo) Save(_ context.Context, a *assignment.Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[a.ID] = a

	return nil
}

func (m *MockAssignmentRepo) GetByID(_ context.Context, id string) (*assignment.Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.store[id]
	if !ok {
		return nil, service.ErrNotFound
	}

	return a, nil
}

func (m *MockAssignmentRepo) GetByDroneID(_ context.Context, droneID string) (*assignment.Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var best *assignment.Assignment
	for _, a := range m.store {
		if a.DroneID != droneID {
			continue
		}

		if a.Status == assignment.AssignmentStatusCompleted || a.Status == assignment.AssignmentStatusFailed {
			continue
		}
		if best == nil || a.CreatedAt.After(best.CreatedAt) {
			best = a
		}
	}

	if best == nil {
		return nil, service.ErrNotFound
	}

	return best, nil
}

func (m *MockAssignmentRepo) UpdateStatus(_ context.Context, id string, status assignment.AssignmentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.store[id]
	if !ok {
		return service.ErrNotFound
	}

	a.Status = status
	m.store[a.ID] = a
	return nil
}
