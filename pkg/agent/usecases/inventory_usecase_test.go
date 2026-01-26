package usecases

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
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

// InventoryUseCaseTestSuite agrupa testes do InventoryUseCase.
type InventoryUseCaseTestSuite struct {
	suite.Suite
	mockPodAnalyzer *mocks.MockAnalyzer[corev1.Pod, reports.PodAnalysis]
	mockTransmitter *mocks.MockTransmitter
	useCase         *InventoryUseCase
	config          *contracts.SyncConfig
}

// SetupTest configura mocks antes de cada teste.
func (s *InventoryUseCaseTestSuite) SetupTest() {
	s.mockPodAnalyzer = new(mocks.MockAnalyzer[corev1.Pod, reports.PodAnalysis])
	s.mockTransmitter = new(mocks.MockTransmitter)

	s.useCase = &InventoryUseCase{
		podAnalyzer: s.mockPodAnalyzer,
		transmitter: s.mockTransmitter,
	}

	s.config = &contracts.SyncConfig{
		SpokeClusterName: "test-cluster",
	}
}

// TestNilConfig testa erro quando config é nil.
func (s *InventoryUseCaseTestSuite) TestNilConfig() {
	// Act
	result, err := s.useCase.Perform(context.Background(), nil)

	// Assert
	s.Error(err)
	s.Equal(agenterrors.ErrNilConfig, err)
	s.Empty(result)

	s.mockPodAnalyzer.AssertNotCalled(s.T(), "Analyze")
	s.mockTransmitter.AssertNotCalled(s.T(), "Transmit")
}

// TestEmptyClusterName testa erro quando cluster name está vazio.
func (s *InventoryUseCaseTestSuite) TestEmptyClusterName() {
	// Arrange
	configWithoutName := &contracts.SyncConfig{
		SpokeClusterName: "",
	}

	// Act
	result, err := s.useCase.Perform(context.Background(), configWithoutName)

	// Assert
	s.Error(err)
	s.Equal(agenterrors.ErrEmptyClusterName, err)
	s.Empty(result)

	s.mockPodAnalyzer.AssertNotCalled(s.T(), "Analyze")
	s.mockTransmitter.AssertNotCalled(s.T(), "Transmit")
}

// TestAnalyzerError testa erro quando analyzer falha.
func (s *InventoryUseCaseTestSuite) TestAnalyzerError() {
	// Arrange
	expectedErr := errors.New("analyzer error")
	s.mockPodAnalyzer.On("Analyze", mock.Anything, s.config).
		Return(reports.PodAnalysis{}, expectedErr)

	// Act
	result, err := s.useCase.Perform(context.Background(), s.config)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Empty(result)

	s.mockPodAnalyzer.AssertExpectations(s.T())
	s.mockTransmitter.AssertNotCalled(s.T(), "Transmit")
}

// TestTransmitterError testa erro quando transmitter falha.
func (s *InventoryUseCaseTestSuite) TestTransmitterError() {
	// Arrange
	podAnalysis := reports.PodAnalysis{
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
	result, err := s.useCase.Perform(context.Background(), s.config)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Empty(result)

	s.mockPodAnalyzer.AssertExpectations(s.T())
	s.mockTransmitter.AssertExpectations(s.T())
}

// TestSuccessfulExecution testa execução bem-sucedida completa.
func (s *InventoryUseCaseTestSuite) TestSuccessfulExecution() {
	// Arrange
	podAnalysis := reports.PodAnalysis{
		TotalPods:   10,
		RunningPods: 8,
		PendingPods: 1,
		FailedPods:  1,
	}

	s.mockPodAnalyzer.On("Analyze", mock.Anything, s.config).
		Return(podAnalysis, nil)
	s.mockTransmitter.On("Transmit", mock.Anything, mock.MatchedBy(func(report reports.ClusterInventoryReport) bool {
		// Validar estrutura do report
		return report.ClusterName == "test-cluster" &&
			report.PodAnalysis != nil &&
			report.PodAnalysis.TotalPods == 10
	}), s.config).Return(nil)
	s.mockTransmitter.On("Name").Return("MockTransmitter")

	// Act
	result, err := s.useCase.Perform(context.Background(), s.config)

	// Assert
	s.NoError(err)
	s.Equal("test-cluster", result.ClusterName)
	s.NotNil(result.PodAnalysis)
	s.Equal(10, result.PodAnalysis.TotalPods)
	s.Equal(8, result.PodAnalysis.RunningPods)
	s.Equal(1, result.PodAnalysis.PendingPods)
	s.Equal(1, result.PodAnalysis.FailedPods)

	s.mockPodAnalyzer.AssertExpectations(s.T())
	s.mockTransmitter.AssertExpectations(s.T())
}

// TestSuite executa a suite de testes.
func TestInventoryUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryUseCaseTestSuite))
}
