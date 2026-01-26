package reports

// NodeAnalysis representa a análise de nodes do cluster.
// Gerado por NodeAnalyzer (NodeCollector + NodeProcessor).
type NodeAnalysis struct {
	// TotalNodes é o número total de nodes no cluster
	TotalNodes int `json:"totalNodes"`

	// ReadyNodes é o número de nodes em estado Ready
	ReadyNodes int `json:"readyNodes"`

	// Nodes contém informações detalhadas de cada node
	Nodes []NodeInfo `json:"nodes"`
}

// NodeInfo contém informações sobre um node individual.
type NodeInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Role   string `json:"role,omitempty"`
}
