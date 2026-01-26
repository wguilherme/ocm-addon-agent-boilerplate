package usecases

import (
	"github.com/totvs/addon-framework-basic/pkg/agent/analyzers"
	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
	"github.com/totvs/addon-framework-basic/pkg/agent/transmitters"
)

var (
	InventoryUseCaseInstance contracts.UseCase[*contracts.SyncConfig, reports.ClusterInventoryReport]
)

func init() {
	InventoryUseCaseInstance = NewInventoryUseCase(
		analyzers.PodAnalyzerInstance,
		transmitters.ConfigMapTransmitterInstance,
	)
}
