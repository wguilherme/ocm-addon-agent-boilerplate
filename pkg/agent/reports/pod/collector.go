package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
)

type podCollector struct {
	namespaces []string // Se vazio, coleta de todos
}

func NewPodCollector() contracts.Collector[corev1.Pod] {
	return &podCollector{namespaces: []string{}}
}

func NewPodCollectorWithNamespaces(namespaces []string) contracts.Collector[corev1.Pod] {
	return &podCollector{namespaces: namespaces}
}

func (c *podCollector) Collect(ctx context.Context, config *contracts.SyncConfig) ([]corev1.Pod, error) {
	if config == nil {
		return nil, agenterrors.ErrNilConfig
	}
	if config.SpokeClient == nil {
		return nil, agenterrors.ErrNilSpokeClient
	}

	klog.V(4).Infof("[PodCollector] Starting collection from spoke cluster '%s'", config.SpokeClusterName)

	var allPods []corev1.Pod

	if len(c.namespaces) > 0 {
		for _, ns := range c.namespaces {
			podList, err := config.SpokeClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				klog.Errorf("[PodCollector] Failed to list pods in namespace '%s': %v", ns, err)
				return nil, err
			}
			allPods = append(allPods, podList.Items...)
		}
	} else {
		podList, err := config.SpokeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("[PodCollector] Failed to list pods from all namespaces: %v", err)
			return nil, err
		}
		allPods = podList.Items
	}

	klog.V(4).Infof("[PodCollector] Successfully collected %d pods", len(allPods))
	return allPods, nil
}

func (c *podCollector) Name() string {
	return "PodCollector"
}
