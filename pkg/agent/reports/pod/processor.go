package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
)

type podProcessor struct{}

func NewPodProcessor() contracts.Processor[corev1.Pod, contracts.PodAnalysis] {
	return &podProcessor{}
}

func (p *podProcessor) Process(ctx context.Context, pods []corev1.Pod, clusterName string) (contracts.PodAnalysis, error) {
	klog.V(4).Infof("[PodProcessor] Processing %d pods for cluster '%s'", len(pods), clusterName)

	analysis := contracts.PodAnalysis{
		TotalPods:   len(pods),
		PodsByPhase: make(map[string]int),
		Pods:        make([]contracts.PodInfo, 0, len(pods)),
	}

	for _, pod := range pods {
		phase := string(pod.Status.Phase)
		analysis.PodsByPhase[phase]++

		switch pod.Status.Phase {
		case corev1.PodRunning:
			analysis.RunningPods++
		case corev1.PodPending:
			analysis.PendingPods++
		case corev1.PodFailed:
			analysis.FailedPods++
		}

		analysis.Pods = append(analysis.Pods, contracts.PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Phase:     phase,
			NodeName:  pod.Spec.NodeName,
		})
	}

	klog.V(4).Infof("[PodProcessor] Analysis complete: Total=%d, Running=%d, Pending=%d, Failed=%d",
		analysis.TotalPods, analysis.RunningPods, analysis.PendingPods, analysis.FailedPods)

	return analysis, nil
}

func (p *podProcessor) Name() string {
	return "PodProcessor"
}
