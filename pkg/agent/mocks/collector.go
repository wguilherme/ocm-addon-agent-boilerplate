package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
)

// MockCollector is a mock implementation of contracts.Collector.
type MockCollector[T any] struct {
	mock.Mock
}

func (m *MockCollector[T]) Collect(ctx context.Context, config *contracts.SyncConfig) ([]T, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockCollector[T]) Name() string {
	args := m.Called()
	return args.String(0)
}
