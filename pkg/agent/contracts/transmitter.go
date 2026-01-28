package contracts

import "context"

// Transmitter envia relatórios para o hub cluster (CRD, ConfigMap, Status).
type Transmitter interface {
	Transmit(ctx context.Context, report ClusterInventoryReport, config *SyncConfig) error
	Name() string
}
