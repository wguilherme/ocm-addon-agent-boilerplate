package reports

// PodAnalysis representa a análise de pods do cluster.
// Gerado por PodAnalyzer (PodCollector + PodProcessor).
type PodAnalysis struct {
	// TotalPods é o número total de pods no cluster
	TotalPods int `json:"totalPods"`

	// RunningPods é o número de pods em estado Running
	RunningPods int `json:"runningPods"`

	// PendingPods é o número de pods em estado Pending
	PendingPods int `json:"pendingPods"`

	// FailedPods é o número de pods em estado Failed
	FailedPods int `json:"failedPods"`

	// PodsByPhase agrupa pods por fase (Running, Pending, Failed, etc.)
	PodsByPhase map[string]int `json:"podsByPhase"`

	// Pods contém informações detalhadas de cada pod
	Pods []PodInfo `json:"pods"`
}

// PodInfo contém informações sobre um pod individual.
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Phase     string `json:"phase"`
	NodeName  string `json:"nodeName,omitempty"`
}
