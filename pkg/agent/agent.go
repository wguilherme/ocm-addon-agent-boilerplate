package agent

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	cmdfactory "open-cluster-management.io/addon-framework/pkg/cmd/factory"
	"open-cluster-management.io/addon-framework/pkg/version"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
	"github.com/totvs/addon-framework-basic/pkg/agent/orchestrators"
	"github.com/totvs/addon-framework-basic/pkg/agent/reports/pod"
	"github.com/totvs/addon-framework-basic/pkg/agent/transmitters"
)

// NewAgentCommand creates the agent subcommand.
func NewAgentCommand(addonName string) *cobra.Command {
	o := NewAgentOptions(addonName)
	cmd := cmdfactory.
		NewControllerCommandConfig(addonName+"-agent", version.Get(), o.RunAgent).
		NewCommand()
	cmd.Use = "agent"
	cmd.Short = "Start the addon agent"

	o.AddFlags(cmd)
	return cmd
}

// RunAgent starts the agent that collects pod info and sends to hub.
func (o *AgentOptions) RunAgent(ctx context.Context, kubeconfig *rest.Config) error {
	klog.Infof("Starting %s agent", o.AddonName)

	// Build spoke client (local cluster)
	spokeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}
	klog.Info("Connected to spoke cluster")

	// Build hub client
	hubRestConfig, err := clientcmd.BuildConfigFromFlags("", o.HubKubeconfigFile)
	if err != nil {
		return err
	}
	hubClient, err := kubernetes.NewForConfig(hubRestConfig)
	if err != nil {
		return err
	}
	klog.Infof("Connected to hub cluster, will report to namespace: %s", o.SpokeClusterName)

	// Create sync configuration
	syncConfig := &contracts.SyncConfig{
		SpokeClusterName: o.SpokeClusterName,
		AddonName:        o.AddonName,
		AddonNamespace:   o.AddonNamespace,
		SpokeClient:      spokeClient,
		HubClient:        hubClient,
	}

	// Initialize sync services (explicit DI)
	syncServices := []contracts.SyncService{
		orchestrators.NewInventoryOrchestrator(
			pod.NewPodAnalyzer(
				pod.NewPodCollector(),
				pod.NewPodProcessor(),
			),
			transmitters.NewConfigMapTransmitter("cluster-inventory-report"),
		),
		// Future: orchestrators.NewMetricsOrchestrator(...),
		// Future: orchestrators.NewEventsOrchestrator(...),
	}

	// Start sync loop
	ticker := time.NewTicker(SyncInterval)
	defer ticker.Stop()

	// Run immediately once, then on ticker
	runAllServices(ctx, syncServices, syncConfig)

	for {
		select {
		case <-ctx.Done():
			klog.Info("Agent shutting down")
			return nil
		case <-ticker.C:
			runAllServices(ctx, syncServices, syncConfig)
		}
	}
}

// runAllServices executes all sync services and logs errors individually.
func runAllServices(ctx context.Context, services []contracts.SyncService, config *contracts.SyncConfig) {
	for _, service := range services {
		if err := service.Sync(ctx, config); err != nil {
			klog.Errorf("Failed to sync %s: %v", service.Name(), err)
		}
	}
}
