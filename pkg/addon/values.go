package addon

import (
	"fmt"
	"os"

	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	addonapiv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

// ManifestConfig holds the configuration for rendering manifest templates.
type ManifestConfig struct {
	KubeConfigSecret string
	ClusterName      string
	Image            string
}

// GetDefaultValues returns the default values for the addon manifests.
// These values are injected into the Go templates.
func GetDefaultValues(cluster *clusterv1.ManagedCluster,
	addon *addonapiv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {

	image := os.Getenv("ADDON_IMAGE")
	if len(image) == 0 {
		image = DefaultAddonImage
	}

	manifestConfig := ManifestConfig{
		KubeConfigSecret: fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
		ClusterName:      cluster.Name,
		Image:            image,
	}

	return addonfactory.StructToValues(manifestConfig), nil
}
