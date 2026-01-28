package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
)

// MockTransmitter is a mock implementation of contracts.Transmitter.
type MockTransmitter struct {
	mock.Mock
}

func (m *MockTransmitter) Transmit(ctx context.Context, report contracts.ClusterInventoryReport, config *contracts.SyncConfig) error {
	args := m.Called(ctx, report, config)
	return args.Error(0)
}

func (m *MockTransmitter) Name() string {
	args := m.Called()
	return args.String(0)
}
