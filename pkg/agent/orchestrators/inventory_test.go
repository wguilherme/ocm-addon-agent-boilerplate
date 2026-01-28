package orchestrators

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
	"github.com/totvs/addon-framework-basic/pkg/agent/mocks"
)

// InventoryOrchestratorTestSuite agrupa testes do InventoryOrchestrator.
type InventoryOrchestratorTestSuite struct {
	suite.Suite
	mockPodAnalyzer *mocks.MockAnalyzer[corev1.Pod, contracts.PodAnalysis]
	mockTransmitter *mocks.MockTransmitter
	orchestrator    *InventoryOrchestrator
	config          *contracts.SyncConfig
}

// SetupTest configura mocks antes de cada teste.
func (s *InventoryOrchestratorTestSuite) SetupTest() {
	s.mockPodAnalyzer = new(mocks.MockAnalyzer[corev1.Pod, contracts.PodAnalysis])
	s.mockTransmitter = new(mocks.MockTransmitter)

	s.orchestrator = &InventoryOrchestrator{
		podAnalyzer: s.mockPodAnalyzer,
		transmitter: s.mockTransmitter,
	}

	s.config = &contracts.SyncConfig{
		SpokeClusterName: "test-cluster",
	}
}

// TestNilConfig testa erro quando config é nil.
func (s *InventoryOrchestratorTestSuite) TestNilConfig() {
	// Act
	err := s.orchestrator.Sync(context.Background(), nil)

	// Assert
	s.Error(err)
	s.Equal(agenterrors.ErrNilConfig, err)

	s.mockPodAnalyzer.AssertNotCalled(s.T(), "Analyze")
	s.mockTransmitter.AssertNotCalled(s.T(), "Transmit")
}

// TestEmptyClusterName testa erro quando cluster name está vazio.
func (s *InventoryOrchestratorTestSuite) TestEmptyClusterName() {
	// Arrange
	configWithoutName := &contracts.SyncConfig{
		SpokeClusterName: "",
	}

	// Act
	err := s.orchestrator.Sync(context.Background(), configWithoutName)

	// Assert
	s.Error(err)
	s.Equal(agenterrors.ErrEmptyClusterName, err)

	s.mockPodAnalyzer.AssertNotCalled(s.T(), "Analyze")
	s.mockTransmitter.AssertNotCalled(s.T(), "Transmit")
}

// TestAnalyzerError testa erro quando analyzer falha.
func (s *InventoryOrchestratorTestSuite) TestAnalyzerError() {
	// Arrange
	expectedErr := errors.New("analyzer error")
	s.mockPodAnalyzer.On("Analyze", mock.Anything, s.config).
		Return(contracts.PodAnalysis{}, expectedErr)

	// Act
	err := s.orchestrator.Sync(context.Background(), s.config)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)

	s.mockPodAnalyzer.AssertExpectations(s.T())
	s.mockTransmitter.AssertNotCalled(s.T(), "Transmit")
}

// TestTransmitterError testa erro quando transmitter falha.
func (s *InventoryOrchestratorTestSuite) TestTransmitterError() {
	// Arrange
	podAnalysis := contracts.PodAnalysis{
		TotalPods:   5,
		RunningPods: 3,
	}

	expectedErr := errors.New("transmission error")

	s.mockPodAnalyzer.On("Analyze", mock.Anything, s.config).
		Return(podAnalysis, nil)
	s.mockTransmitter.On("Transmit", mock.Anything, mock.Anything, s.config).
		Return(expectedErr)
	s.mockTransmitter.On("Name").Return("MockTransmitter")

	// Act
	err := s.orchestrator.Sync(context.Background(), s.config)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)

	s.mockPodAnalyzer.AssertExpectations(s.T())
	s.mockTransmitter.AssertExpectations(s.T())
}

// TestSuccessfulExecution testa execução bem-sucedida completa.
func (s *InventoryOrchestratorTestSuite) TestSuccessfulExecution() {
	// Arrange
	podAnalysis := contracts.PodAnalysis{
		TotalPods:   10,
		RunningPods: 8,
		PendingPods: 1,
		FailedPods:  1,
	}

	s.mockPodAnalyzer.On("Analyze", mock.Anything, s.config).
		Return(podAnalysis, nil)
	s.mockTransmitter.On("Transmit", mock.Anything, mock.MatchedBy(func(report contracts.ClusterInventoryReport) bool {
		// Validar estrutura do report
		return report.ClusterName == "test-cluster" &&
			report.PodAnalysis != nil &&
			report.PodAnalysis.TotalPods == 10
	}), s.config).Return(nil)
	s.mockTransmitter.On("Name").Return("MockTransmitter")

	// Act
	err := s.orchestrator.Sync(context.Background(), s.config)

	// Assert
	s.NoError(err)

	s.mockPodAnalyzer.AssertExpectations(s.T())
	s.mockTransmitter.AssertExpectations(s.T())
}

// TestSuite executa a suite de testes.
func TestInventoryOrchestratorTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryOrchestratorTestSuite))
}
