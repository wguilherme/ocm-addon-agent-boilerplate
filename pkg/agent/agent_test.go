package agent

import (
	"testing"
)

func TestNewAgentOptions(t *testing.T) {
	opts := NewAgentOptions("test-addon")

	if opts.AddonName != "test-addon" {
		t.Errorf("AddonName = %s, want test-addon", opts.AddonName)
	}

	if opts.HubKubeconfigFile != "" {
		t.Errorf("HubKubeconfigFile should be empty by default")
	}

	if opts.SpokeClusterName != "" {
		t.Errorf("SpokeClusterName should be empty by default")
	}

	if opts.AddonNamespace != "" {
		t.Errorf("AddonNamespace should be empty by default")
	}
}
