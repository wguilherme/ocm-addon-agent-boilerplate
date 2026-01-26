package analyzers

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/totvs/addon-framework-basic/pkg/agent/analyzers/pod"
	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

var (
	PodAnalyzerInstance contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
)

func init() {
	PodAnalyzerInstance = pod.NewPodAnalyzer(
		pod.NewPodCollector(),
		pod.NewPodProcessor(),
	)
}
