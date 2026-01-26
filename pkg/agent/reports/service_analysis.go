package reports

// ServiceAnalysis representa a análise de services do cluster.
// Gerado por ServiceAnalyzer (ServiceCollector + ServiceProcessor).
type ServiceAnalysis struct {
	// TotalServices é o número total de services no cluster
	TotalServices int `json:"totalServices"`

	// ServicesByType agrupa services por tipo (ClusterIP, NodePort, LoadBalancer, etc.)
	ServicesByType map[string]int `json:"servicesByType"`

	// ExternalServices é o número de services externos (LoadBalancer ou NodePort)
	ExternalServices int `json:"externalServices"`

	// Services contém informações detalhadas de cada service
	Services []ServiceInfo `json:"services"`
}

// ServiceInfo contém informações sobre um service individual.
type ServiceInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIP,omitempty"`
}
