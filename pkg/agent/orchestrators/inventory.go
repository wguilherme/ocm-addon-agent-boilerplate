package orchestrators

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
)

// InventoryOrchestrator coordinates multiple analyzers to build a cluster inventory report.
type InventoryOrchestrator struct {
	podAnalyzer contracts.Analyzer[corev1.Pod, contracts.PodAnalysis]
	// Future: nodeAnalyzer, serviceAnalyzer, ingressAnalyzer
	transmitter contracts.Transmitter
}

// NewInventoryOrchestrator creates a new inventory synchronization orchestrator.
func NewInventoryOrchestrator(
	podAnalyzer contracts.Analyzer[corev1.Pod, contracts.PodAnalysis],
	transmitter contracts.Transmitter,
) contracts.SyncService {
	if podAnalyzer == nil {
		panic("podAnalyzer cannot be nil")
	}
	if transmitter == nil {
		panic("transmitter cannot be nil")
	}

	return &InventoryOrchestrator{
		podAnalyzer: podAnalyzer,
		transmitter: transmitter,
	}
}

// Name returns the orchestrator name for logging.
func (o *InventoryOrchestrator) Name() string {
	return "InventoryOrchestrator"
}

// Sync executes the inventory synchronization.
func (o *InventoryOrchestrator) Sync(ctx context.Context, config *contracts.SyncConfig) error {
	if config == nil {
		return agenterrors.ErrNilConfig
	}
	if config.SpokeClusterName == "" {
		return agenterrors.ErrEmptyClusterName
	}

	klog.Infof("[InventoryOrchestrator] Starting inventory analysis for cluster '%s'", config.SpokeClusterName)
	startTime := time.Now()

	report := contracts.ClusterInventoryReport{
		ClusterName: config.SpokeClusterName,
		Timestamp:   time.Now().UTC(),
	}

	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)

	// Pod analysis (parallel execution ready for future analyzers)
	g.Go(func() error {
		klog.V(4).Infof("[InventoryOrchestrator] Starting PodAnalyzer...")
		podAnalysis, err := o.podAnalyzer.Analyze(ctx, config)
		if err != nil {
			klog.Errorf("[InventoryOrchestrator] PodAnalyzer failed: %v", err)
			return err
		}

		mu.Lock()
		report.PodAnalysis = &podAnalysis
		mu.Unlock()

		klog.V(4).Infof("[InventoryOrchestrator] PodAnalyzer completed: %d total pods", podAnalysis.TotalPods)
		return nil
	})

	if err := g.Wait(); err != nil {
		klog.Errorf("[InventoryOrchestrator] One or more analyzers failed: %v", err)
		return err
	}

	duration := time.Since(startTime)
	klog.Infof("[InventoryOrchestrator] All analyzers completed in %v", duration)

	klog.V(4).Infof("[InventoryOrchestrator] Transmitting report via %s...", o.transmitter.Name())
	if err := o.transmitter.Transmit(ctx, report, config); err != nil {
		klog.Errorf("[InventoryOrchestrator] Transmission failed: %v", err)
		return err
	}

	klog.Infof("[InventoryOrchestrator] Inventory analysis completed successfully for cluster '%s'", config.SpokeClusterName)
	return nil
}
