package usecases

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	agenterrors "github.com/totvs/addon-framework-basic/pkg/agent/errors"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports"
)

// InventoryUseCase → orquestra múltiplos analyzers e transmite relatório agregado.
type InventoryUseCase struct {
	podAnalyzer contracts.Analyzer[corev1.Pod, reports.PodAnalysis]
	// TODO: adicionar outros analyzers (service, ingress, node)
	transmitter contracts.Transmitter
}

func NewInventoryUseCase(
	podAnalyzer contracts.Analyzer[corev1.Pod, reports.PodAnalysis],
	transmitter contracts.Transmitter,
) contracts.UseCase[*contracts.SyncConfig, reports.ClusterInventoryReport] {
	if podAnalyzer == nil {
		panic("podAnalyzer cannot be nil")
	}
	if transmitter == nil {
		panic("transmitter cannot be nil")
	}

	return &InventoryUseCase{
		podAnalyzer: podAnalyzer,
		transmitter: transmitter,
	}
}

func (u *InventoryUseCase) Perform(ctx context.Context, config *contracts.SyncConfig) (reports.ClusterInventoryReport, error) {
	if config == nil {
		return reports.ClusterInventoryReport{}, agenterrors.ErrNilConfig
	}
	if config.SpokeClusterName == "" {
		return reports.ClusterInventoryReport{}, agenterrors.ErrEmptyClusterName
	}

	klog.Infof("[InventoryUseCase] Starting inventory analysis for cluster '%s'", config.SpokeClusterName)
	startTime := time.Now()

	report := reports.ClusterInventoryReport{
		ClusterName: config.SpokeClusterName,
		Timestamp:   time.Now().UTC(),
	}

	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		klog.V(4).Infof("[InventoryUseCase] Starting PodAnalyzer...")
		podAnalysis, err := u.podAnalyzer.Analyze(ctx, config)
		if err != nil {
			klog.Errorf("[InventoryUseCase] PodAnalyzer failed: %v", err)
			return err
		}

		mu.Lock()
		report.PodAnalysis = &podAnalysis
		mu.Unlock()

		klog.V(4).Infof("[InventoryUseCase] PodAnalyzer completed: %d total pods", podAnalysis.TotalPods)
		return nil
	})

	if err := g.Wait(); err != nil {
		klog.Errorf("[InventoryUseCase] One or more analyzers failed: %v", err)
		return reports.ClusterInventoryReport{}, err
	}

	duration := time.Since(startTime)
	klog.Infof("[InventoryUseCase] All analyzers completed in %v", duration)

	klog.V(4).Infof("[InventoryUseCase] Transmitting report via %s...", u.transmitter.Name())
	if err := u.transmitter.Transmit(ctx, report, config); err != nil {
		klog.Errorf("[InventoryUseCase] Transmission failed: %v", err)
		return reports.ClusterInventoryReport{}, err
	}

	klog.Infof("[InventoryUseCase] Inventory analysis completed successfully for cluster '%s'", config.SpokeClusterName)
	return report, nil
}
