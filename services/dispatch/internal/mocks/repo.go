package mocks

import (
	"context"
	"hive/services/dispatch/internal/service"
	"sync"

	"hive/services/dispatch/internal/domain"
)

type MockAssignmentRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.Assignment
}

func NewMockAssignmentRepo() *MockAssignmentRepo {
	return &MockAssignmentRepo{
		store: make(map[string]*domain.Assignment),
	}
}

func (m *MockAssignmentRepo) Save(_ context.Context, a *domain.Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[a.ID] = a

	return nil
}

func (m *MockAssignmentRepo) Get(_ context.Context, id string) (*domain.Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.store[id]
	if !ok {
		return nil, service.ErrNotFound
	}

	return a, nil
}

func (m *MockAssignmentRepo) Update(_ context.Context, a *domain.Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.store[a.ID]
	if !ok {
		return service.ErrNotFound
	}

	m.store[a.ID] = a
	return nil
}
