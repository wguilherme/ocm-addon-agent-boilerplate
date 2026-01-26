package pod

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
)

type podAnalyzer struct {
	collector Collector
	processor Processor
}

func NewPodAnalyzer(
	collector Collector,
	processor Processor,
) Analyzer {
	if collector == nil {
		panic("collector cannot be nil")
	}
	if processor == nil {
		panic("processor cannot be nil")
	}

	return &podAnalyzer{
		collector: collector,
		processor: processor,
	}
}

func (a *podAnalyzer) Analyze(ctx context.Context, config *contracts.SyncConfig) (PodOutput, error) {
	klog.V(4).Infof("[PodAnalyzer] Starting analysis for cluster '%s'", config.SpokeClusterName)

	pods, err := a.collector.Collect(ctx, config)
	if err != nil {
		return PodOutput{}, agenterrors.NewCollectionError(a.Name(), err)
	}

	analysis, err := a.processor.Process(ctx, pods, config.SpokeClusterName)
	if err != nil {
		return PodOutput{}, agenterrors.NewProcessingError(a.processor.Name(), err)
	}

	klog.V(4).Infof("[PodAnalyzer] Analysis completed successfully")
	return analysis, nil
}

func (a *podAnalyzer) Name() string {
	return "PodAnalyzer"
}
