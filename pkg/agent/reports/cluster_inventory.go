package reports

import "time"

// ClusterInventoryReport é o report final agregado enviado para o hub.
// Contém múltiplas seções de análise, cada uma preenchida por um Analyzer específico.
//
// Exemplo de uso:
//
//	report := ClusterInventoryReport{
//	    ClusterName: "spoke1-sftm",
//	    Timestamp:   time.Now().UTC(),
//	    PodAnalysis: &PodAnalysis{...},      // Preenchido por PodAnalyzer
//	    ServiceAnalysis: &ServiceAnalysis{...}, // Preenchido por ServiceAnalyzer
//	}
type ClusterInventoryReport struct {
	// ClusterName identifica o cluster spoke de origem
	ClusterName string `json:"clusterName"`

	// Timestamp indica quando o report foi gerado
	Timestamp time.Time `json:"timestamp"`

	// PodAnalysis contém análise de pods (opcional)
	// Preenchido por PodAnalyzer
	PodAnalysis *PodAnalysis `json:"podAnalysis,omitempty"`

	// ServiceAnalysis contém análise de services (opcional)
	// Preenchido por ServiceAnalyzer
	ServiceAnalysis *ServiceAnalysis `json:"serviceAnalysis,omitempty"`

	// IngressAnalysis contém análise de ingresses (opcional)
	// Preenchido por IngressAnalyzer
	IngressAnalysis *IngressAnalysis `json:"ingressAnalysis,omitempty"`

	// NodeAnalysis contém análise de nodes (opcional)
	// Preenchido por NodeAnalyzer
	NodeAnalysis *NodeAnalysis `json:"nodeAnalysis,omitempty"`
}
