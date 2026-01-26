# Shared variables for automation modules
# Override via .env file or environment variables

# Addon configuration
ADDON_NAME ?= ocm-addon-boilerplate
IMAGE ?= $(ADDON_NAME):latest

# Number of spoke clusters to create (override: SPOKE_COUNT in .env)
SPOKE_COUNT ?= $(or $(TEST_AUTOMATION_SPOKE_COUNT),1)

# Target spoke for operations (override: SPOKE in .env)
SPOKE ?= $(or $(TEST_AUTOMATION_SPOKE_NAME),spoke1)

# Cluster name for operations (override: CLUSTER in .env)
CLUSTER ?= $(or $(TEST_AUTOMATION_CLUSTER_NAME),$(SPOKE))

# Directory for exported Kind kubeconfigs
KUBECONFIG_KIND_EXTRACT_CONFIG_PATH ?= ~/.kube/local/$(ADDON_NAME)

# Helper to expand ~ in paths
KUBECONFIG_PATH_EXPANDED = $(shell echo $(KUBECONFIG_KIND_EXTRACT_CONFIG_PATH) | sed "s|^~|$$HOME|")

# Kubeconfig paths - Hub is static, Spoke is dynamic based on SPOKE variable
HUB_KUBECONFIG    := $(KUBECONFIG_KIND_EXTRACT_CONFIG_PATH)/config.hub
SPOKE_KUBECONFIG  := $(KUBECONFIG_KIND_EXTRACT_CONFIG_PATH)/config.$(SPOKE)

# Kind cluster names
KIND_HUB   ?= hub
KIND_SPOKE ?= $(SPOKE)

# Tool versions
KIND_VERSION      ?= v0.20.0
KUBECTL_VERSION   ?= v1.28.2
CLUSTERADM_VERSION ?= latest
