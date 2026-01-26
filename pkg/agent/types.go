package agent

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// PodReport is the structure sent to the hub with pod information.
// This structure is extensible - add more fields as needed.
type PodReport struct {
	ClusterName string    `json:"clusterName"`
	Timestamp   time.Time `json:"timestamp"`
	TotalPods   int       `json:"totalPods"`
	Pods        []PodInfo `json:"pods"`
}

// PodInfo contains information about a single pod.
// Add more fields as needed for your use case.
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	NodeName  string `json:"nodeName,omitempty"`
}

// BuildPodReport creates a PodReport from a list of pods.
func BuildPodReport(clusterName string, pods []corev1.Pod) PodReport {
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
