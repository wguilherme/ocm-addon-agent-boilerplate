package strategies

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

// InventoryStrategy → pattern Collector + Processor + Analyzer.
// Fluxo: Analyzers paralelos → ClusterInventoryReport → Transmitter → Hub
type InventoryStrategy struct {
	useCase contracts.UseCase[*SyncConfig, reports.ClusterInventoryReport]
}

func NewInventoryStrategy(
	useCase contracts.UseCase[*SyncConfig, reports.ClusterInventoryReport],
) SyncStrategy {
	if useCase == nil {
		panic("useCase cannot be nil")
	}

	return &InventoryStrategy{
		useCase: useCase,
	}
}

func (s *InventoryStrategy) Name() string {
	return "InventoryStrategy"
}

func (s *InventoryStrategy) Sync(ctx context.Context, config *SyncConfig) error {
	klog.Infof("[InventoryStrategy] Starting inventory sync for cluster '%s'", config.SpokeClusterName)

	_, err := s.useCase.Perform(ctx, config)
	if err != nil {
		klog.Errorf("[InventoryStrategy] Inventory sync failed: %v", err)
		return err
	}

	klog.Infof("[InventoryStrategy] Inventory sync completed successfully")
	return nil
}
