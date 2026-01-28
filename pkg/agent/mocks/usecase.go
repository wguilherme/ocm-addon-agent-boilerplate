package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockUseCase is a mock implementation of contracts.UseCase.
// Deprecated: UseCase pattern is no longer used.
type MockUseCase[I any, O any] struct {
	mock.Mock
}

func (m *MockUseCase[I, O]) Perform(ctx context.Context, input I) (O, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(O), args.Error(1)
}
