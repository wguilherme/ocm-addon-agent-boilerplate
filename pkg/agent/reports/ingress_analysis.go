package reports

// IngressAnalysis representa a análise de ingresses do cluster.
// Gerado por IngressAnalyzer (IngressCollector + IngressProcessor).
type IngressAnalysis struct {
	// TotalIngresses é o número total de ingresses no cluster
	TotalIngresses int `json:"totalIngresses"`

	// IngressesWithTLS é o número de ingresses com TLS configurado
	IngressesWithTLS int `json:"ingressesWithTLS"`

	// Ingresses contém informações detalhadas de cada ingress
	Ingresses []IngressInfo `json:"ingresses"`
}

// IngressInfo contém informações sobre um ingress individual.
type IngressInfo struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	Hosts        []string `json:"hosts"`
	TLS          bool     `json:"tls"`
	IngressClass string   `json:"ingressClass,omitempty"`
}
