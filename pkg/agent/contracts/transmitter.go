package contracts

import (
	"context"

	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

// Transmitter envia relatórios para o hub cluster (CRD, ConfigMap, Status).
type Transmitter interface {
	Transmit(ctx context.Context, report reports.ClusterInventoryReport, config *SyncConfig) error
	Name() string
}
