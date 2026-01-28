package contracts

import "time"

// ClusterInventoryReport é o report final agregado enviado para o hub.
// Contém análise de recursos do cluster.
//
// Exemplo de uso:
//
//	report := ClusterInventoryReport{
//	    ClusterName: "spoke1-sftm",
//	    Timestamp:   time.Now().UTC(),
//	    PodAnalysis: &PodAnalysis{...}, // Preenchido por PodAnalyzer
//	}
type ClusterInventoryReport struct {
	// ClusterName identifica o cluster spoke de origem
	ClusterName string `json:"clusterName"`

	// Timestamp indica quando o report foi gerado
	Timestamp time.Time `json:"timestamp"`

	// PodAnalysis contém análise de pods (opcional)
	// Preenchido por PodAnalyzer
	PodAnalysis *PodAnalysis `json:"podAnalysis,omitempty"`
}

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
