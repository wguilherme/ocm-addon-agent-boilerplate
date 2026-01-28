package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockProcessor is a mock implementation of contracts.Processor.
type MockProcessor[T any, R any] struct {
	mock.Mock
}

func (m *MockProcessor[T, R]) Process(ctx context.Context, data []T, clusterName string) (R, error) {
	args := m.Called(ctx, data, clusterName)
	return args.Get(0).(R), args.Error(1)
}

func (m *MockProcessor[T, R]) Name() string {
	args := m.Called()
	return args.String(0)
}
