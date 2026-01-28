package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
)

// MockAnalyzer is a mock implementation of contracts.Analyzer.
type MockAnalyzer[T any, R any] struct {
	mock.Mock
}

func (m *MockAnalyzer[T, R]) Analyze(ctx context.Context, config *contracts.SyncConfig) (R, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(R), args.Error(1)
}

func (m *MockAnalyzer[T, R]) Name() string {
	args := m.Called()
	return args.String(0)
}
