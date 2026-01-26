package pod_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"

	"github.com/totvs/addon-framework-basic/pkg/agent/analyzers/pod"
	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
	"github.com/totvs/addon-framework-basic/pkg/agent/mocks"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

// PodAnalyzerTestSuite agrupa testes do PodAnalyzer.
type PodAnalyzerTestSuite struct {
	suite.Suite
	mockCollector *mocks.MockCollector[corev1.Pod]
	mockProcessor *mocks.MockProcessor[corev1.Pod, reports.PodAnalysis]
	analyzer      contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
	config        *contracts.SyncConfig
}

// SetupTest configura mocks antes de cada teste.
func (s *PodAnalyzerTestSuite) SetupTest() {
	s.mockCollector = new(mocks.MockCollector[corev1.Pod])
	s.mockProcessor = new(mocks.MockProcessor[corev1.Pod, reports.PodAnalysis])
	s.analyzer = pod.NewPodAnalyzer(s.mockCollector, s.mockProcessor)
	s.config = &contracts.SyncConfig{
		SpokeClusterName: "test-cluster",
	}
}

// TestCollectionError testa erro na coleta.
func (s *PodAnalyzerTestSuite) TestCollectionError() {
	// Arrange
	expectedErr := errors.New("k8s api error")
	s.mockCollector.On("Collect", mock.Anything, s.config).Return(nil, expectedErr)

	// Act
	result, err := s.analyzer.Analyze(context.Background(), s.config)

	// Assert
	s.Error(err)
	s.Empty(result)

	// Verificar que é erro de coleta
	var collectionErr *agenterrors.CollectionError
	s.True(errors.As(err, &collectionErr))
	s.Equal("PodAnalyzer", collectionErr.AnalyzerName)
	s.Equal(expectedErr, errors.Unwrap(err))

	s.mockCollector.AssertExpectations(s.T())
	s.mockProcessor.AssertNotCalled(s.T(), "Process")
}

// TestProcessingError testa erro no processamento.
func (s *PodAnalyzerTestSuite) TestProcessingError() {
	// Arrange
	pods := []corev1.Pod{
		{Status: corev1.PodStatus{Phase: corev1.PodRunning}},
	}
	expectedErr := errors.New("processing error")

	s.mockCollector.On("Collect", mock.Anything, s.config).Return(pods, nil)
	s.mockProcessor.On("Process", mock.Anything, pods, "test-cluster").
		Return(reports.PodAnalysis{}, expectedErr)
	s.mockProcessor.On("Name").Return("PodProcessor")

	// Act
	result, err := s.analyzer.Analyze(context.Background(), s.config)

	// Assert
	s.Error(err)
	s.Empty(result)

	// Verificar que é erro de processamento
	var processingErr *agenterrors.ProcessingError
	s.True(errors.As(err, &processingErr))
	s.Equal("PodProcessor", processingErr.ProcessorName)

	s.mockCollector.AssertExpectations(s.T())
	s.mockProcessor.AssertExpectations(s.T())
}

// TestSuccessfulAnalysis testa análise bem-sucedida.
func (s *PodAnalyzerTestSuite) TestSuccessfulAnalysis() {
	// Arrange
	pods := []corev1.Pod{
		{Status: corev1.PodStatus{Phase: corev1.PodRunning}},
		{Status: corev1.PodStatus{Phase: corev1.PodPending}},
	}

	expectedAnalysis := reports.PodAnalysis{
		TotalPods:   2,
		RunningPods: 1,
		PendingPods: 1,
	}

	s.mockCollector.On("Collect", mock.Anything, s.config).Return(pods, nil)
	s.mockProcessor.On("Process", mock.Anything, pods, "test-cluster").
		Return(expectedAnalysis, nil)

	// Act
	result, err := s.analyzer.Analyze(context.Background(), s.config)

	// Assert
	s.NoError(err)
	s.Equal(expectedAnalysis.TotalPods, result.TotalPods)
	s.Equal(expectedAnalysis.RunningPods, result.RunningPods)
	s.Equal(expectedAnalysis.PendingPods, result.PendingPods)

	s.mockCollector.AssertExpectations(s.T())
	s.mockProcessor.AssertExpectations(s.T())
}

// TestPodProcessor testa processador isoladamente.
func TestPodProcessor(t *testing.T) {
	processor := pod.NewPodProcessor()

	pods := []corev1.Pod{
		{
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
			Spec:   corev1.PodSpec{NodeName: "node1"},
		},
		{
			Status: corev1.PodStatus{Phase: corev1.PodPending},
			Spec:   corev1.PodSpec{NodeName: "node2"},
		},
		{
			Status: corev1.PodStatus{Phase: corev1.PodFailed},
		},
	}

	analysis, err := processor.Process(context.Background(), pods, "test-cluster")

	assert.NoError(t, err)
	assert.Equal(t, 3, analysis.TotalPods)
	assert.Equal(t, 1, analysis.RunningPods)
	assert.Equal(t, 1, analysis.PendingPods)
	assert.Equal(t, 1, analysis.FailedPods)
	assert.Equal(t, 3, len(analysis.Pods))
	assert.Equal(t, 1, analysis.PodsByPhase["Running"])
	assert.Equal(t, 1, analysis.PodsByPhase["Pending"])
	assert.Equal(t, 1, analysis.PodsByPhase["Failed"])
}

// TestSuite executa a suite de testes.
func TestPodAnalyzerTestSuite(t *testing.T) {
	suite.Run(t, new(PodAnalyzerTestSuite))
}
