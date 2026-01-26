package addon

import (
	"fmt"

	"open-cluster-management.io/addon-framework/pkg/agent"
	workapiv1 "open-cluster-management.io/api/work/v1"
)

// AgentHealthProber returns the health prober configuration for the addon.
// Uses WorkProber with FeedbackRules to demonstrate Strategy 5: Work Status Feedback.
// This extracts readyReplicas and availableReplicas from the agent deployment.
func AgentHealthProber() *agent.HealthProber {
	return &agent.HealthProber{
		Type: agent.HealthProberTypeWork,
		WorkProber: &agent.WorkHealthProber{
			ProbeFields: []agent.ProbeField{
				{
					ResourceIdentifier: workapiv1.ResourceIdentifier{
						Group:     "apps",
						Resource:  "deployments",
						Name:      AddonName + "-agent",
						Namespace: InstallationNamespace,
					},
					ProbeRules: []workapiv1.FeedbackRule{
						{
							Type: workapiv1.JSONPathsType,
							JsonPaths: []workapiv1.JsonPath{
								{Name: "readyReplicas", Path: ".status.readyReplicas"},
								{Name: "availableReplicas", Path: ".status.availableReplicas"},
								{Name: "replicas", Path: ".status.replicas"},
							},
						},
					},
				},
			},
			HealthCheck: func(identifier workapiv1.ResourceIdentifier, result workapiv1.StatusFeedbackResult) error {
				if identifier.Name != AddonName+"-agent" {
					return fmt.Errorf("unexpected resource: %s", identifier.Name)
				}
				for _, value := range result.Values {
					if value.Name == "readyReplicas" && value.Value.Integer != nil && *value.Value.Integer >= 1 {
						return nil
					}
				}
				return fmt.Errorf("%s agent is not ready", AddonName)
			},
		},
	}
}
