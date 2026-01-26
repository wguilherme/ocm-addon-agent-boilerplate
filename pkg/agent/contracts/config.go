package contracts

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// SyncConfig holds the configuration and clients needed for sync operations.
type SyncConfig struct {
	// SpokeClusterName is the name of the spoke cluster.
	SpokeClusterName string
	// AddonName is the name of the addon.
	AddonName string
	// AddonNamespace is the namespace where the addon is installed.
	AddonNamespace string

	// SpokeClient is the Kubernetes client for the spoke cluster.
	SpokeClient kubernetes.Interface
	// HubClient is the Kubernetes client for the hub cluster.
	HubClient kubernetes.Interface
	// SpokeDynamicClient is the dynamic client for the spoke cluster.
	SpokeDynamicClient dynamic.Interface
	// HubDynamicClient is the dynamic client for the hub cluster.
	HubDynamicClient dynamic.Interface
}
