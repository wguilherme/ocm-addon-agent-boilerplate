package agent

import (
	"time"

	"github.com/spf13/cobra"
)

const (
	// PodReportConfigMapName is the name of the ConfigMap storing pod reports.
	PodReportConfigMapName = "pod-report"
	// SyncInterval is the interval between sync operations.
	SyncInterval = 60 * time.Second
)

// AgentOptions defines the flags for the agent.
type AgentOptions struct {
	HubKubeconfigFile string
	SpokeClusterName  string
	AddonName         string
	AddonNamespace    string
}

// NewAgentOptions returns the flags with default values.
func NewAgentOptions(addonName string) *AgentOptions {
	return &AgentOptions{AddonName: addonName}
}

// AddFlags registers the agent flags.
func (o *AgentOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVar(&o.HubKubeconfigFile, "hub-kubeconfig", o.HubKubeconfigFile,
		"Location of kubeconfig file to connect to hub cluster.")
	flags.StringVar(&o.SpokeClusterName, "cluster-name", o.SpokeClusterName,
		"Name of the spoke cluster.")
	flags.StringVar(&o.AddonNamespace, "addon-namespace", o.AddonNamespace,
		"Installation namespace of addon.")
	flags.StringVar(&o.AddonName, "addon-name", o.AddonName,
		"Name of the addon.")
}
