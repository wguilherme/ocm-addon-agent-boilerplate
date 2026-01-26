package strategies

import (
	"context"
	"encoding/json"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	// PodReportConfigMapName is the name of the ConfigMap storing pod reports.
	PodReportConfigMapName = "pod-report"
)

// CollectorStrategy implements a strategy that collects pod information
// from the spoke cluster and sends it to the hub via ConfigMap.
type CollectorStrategy struct{}

// NewCollectorStrategy creates a new Collector strategy.
func NewCollectorStrategy() *CollectorStrategy {
	return &CollectorStrategy{}
}

// Name returns the strategy name.
func (s *CollectorStrategy) Name() string {
	return "Collector"
}

// Sync collects pods from spoke and sends report to hub via ConfigMap.
func (s *CollectorStrategy) Sync(ctx context.Context, config *SyncConfig) error {
	klog.V(4).Info("Collecting pod report")

	// List all pods in the spoke cluster
	podList, err := config.SpokeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Build pod report
	report := buildPodReport(config.SpokeClusterName, podList.Items)

	// Serialize to JSON
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return err
	}

	// Create or update ConfigMap in hub
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PodReportConfigMapName,
			Namespace: config.SpokeClusterName,
			Labels: map[string]string{
				"app": config.AddonName,
			},
		},
		Data: map[string]string{
			"report": string(reportJSON),
		},
	}

	existing, err := config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Get(ctx, PodReportConfigMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Create(ctx, configMap, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		klog.Infof("Created pod report with %d pods", len(report.Pods))
		return nil
	}
	if err != nil {
		return err
	}

	// Update existing
	configMap.ResourceVersion = existing.ResourceVersion
	_, err = config.HubClient.CoreV1().ConfigMaps(config.SpokeClusterName).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	klog.Infof("Updated pod report with %d pods", len(report.Pods))
	return nil
}

// PodReport is the structure sent to the hub with pod information.
type PodReport struct {
	ClusterName string    `json:"clusterName"`
	Timestamp   time.Time `json:"timestamp"`
	TotalPods   int       `json:"totalPods"`
	Pods        []PodInfo `json:"pods"`
}

// PodInfo contains information about a single pod.
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	NodeName  string `json:"nodeName,omitempty"`
}

// buildPodReport creates a PodReport from a list of pods.
func buildPodReport(clusterName string, pods []corev1.Pod) PodReport {
	podInfos := make([]PodInfo, 0, len(pods))
	for _, pod := range pods {
		podInfos = append(podInfos, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			NodeName:  pod.Spec.NodeName,
		})
	}

	return PodReport{
		ClusterName: clusterName,
		Timestamp:   time.Now().UTC(),
		TotalPods:   len(pods),
		Pods:        podInfos,
	}
}
