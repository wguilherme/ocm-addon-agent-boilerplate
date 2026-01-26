package addon

import (
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"

	"github.com/totvs/addon-framework-basic/pkg/hub"
)

// NewRegistrationOption returns the registration option for the addon agent.
// This enables the agent to get a kubeconfig to communicate with the hub.
func NewRegistrationOption(kubeConfig *rest.Config, addonName, agentName string) *agent.RegistrationOption {
	return &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(addonName, agentName),
		CSRApproveCheck:   utils.DefaultCSRApprover(agentName),
		PermissionConfig:  hub.AddonRBAC(kubeConfig),
	}
}
