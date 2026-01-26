package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

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

type MockTransmitter struct {
	mock.Mock
}

func (m *MockTransmitter) Transmit(ctx context.Context, report reports.ClusterInventoryReport, config *contracts.SyncConfig) error {
	args := m.Called(ctx, report, config)
	return args.Error(0)
}

func (m *MockTransmitter) Name() string {
	args := m.Called()
	return args.String(0)
}

type MockUseCase[I any, O any] struct {
	mock.Mock
}

func (m *MockUseCase[I, O]) Perform(ctx context.Context, input I) (O, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(O), args.Error(1)
}
