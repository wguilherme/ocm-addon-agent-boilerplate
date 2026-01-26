// Package addon provides factory functions for the OCM addon.
package addon

import (
	"embed"
)

const (
	// AddonName is the name of the addon.
	AddonName = "ocm-addon-boilerplate"
	// DefaultAddonImage is the default image for the addon agent.
	DefaultAddonImage = "ocm-addon-boilerplate:latest"
	// InstallationNamespace is the namespace where the addon agent is installed.
	InstallationNamespace = "open-cluster-management-agent-addon"
)

//go:embed manifests
//go:embed manifests/templates
var FS embed.FS
