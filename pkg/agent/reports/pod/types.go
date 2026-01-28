package pod

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
)

type (
	PodInput = corev1.Pod
	PodOutput = contracts.PodAnalysis
	Collector = contracts.Collector[PodInput]
	Processor = contracts.Processor[PodInput, PodOutput]
	Analyzer = contracts.Analyzer[PodInput, PodOutput]
)
