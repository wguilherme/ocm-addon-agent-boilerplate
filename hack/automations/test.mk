SHELL := /bin/bash

TEST_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(abspath $(TEST_MK_DIR)/../..)

include $(TEST_MK_DIR)variables.mk

.PHONY: unit e2e check

##@ Testing

unit:
	@echo "Running unit tests..."
	@cd $(PROJECT_ROOT) && go test ./pkg/... -v -cover
	@echo "Unit tests complete"

e2e:
	@echo "Running e2e tests..."
	@cd $(PROJECT_ROOT) && go test ./test/e2e/... -v -timeout 10m
	@echo "E2E tests complete"

##@ Diagnostics

check:
	@$(TEST_MK_DIR)scripts/menu.sh $(CLUSTER)

# NOTE: Comments starting with "# Lista" or "# Verifica" are parsed by scripts/menu.sh
# Do not remove or modify them without updating the menu script

# Hub Cluster Checks

# Lista todos os ManagedClusters registrados
check-clusters:
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get managedclusters

# Verifica status do ManagedCluster especifico
check-cluster-status:
	@if [ -z "$(CLUSTER)" ]; then \
		echo "Error: CLUSTER variable not set. Usage: make test-check-cluster-status CLUSTER=spoke1"; \
		exit 1; \
	fi
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get managedcluster $(CLUSTER) -o wide
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get managedcluster $(CLUSTER) -o json | jq '.status.conditions[] | {type, status, reason, message}'

# Lista todos os ManagedClusterAddons
check-addons:
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get managedclusteraddons -A

# Verifica status do addon no cluster
check-addon-status:
	@if [ -z "$(CLUSTER)" ]; then \
		echo "Error: CLUSTER variable not set. Usage: make test-check-addon-status CLUSTER=spoke1"; \
		exit 1; \
	fi
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get managedclusteraddon $(ADDON_NAME) -n $(CLUSTER) -o yaml

# Lista ManifestWorks criados pelo addon
check-manifestworks:
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get manifestwork -A | grep $(ADDON_NAME)

# Verifica o ConfigMap pod-report no hub
check-report:
	@if [ -z "$(CLUSTER)" ]; then \
		echo "Error: CLUSTER variable not set. Usage: make test-check-report CLUSTER=spoke1"; \
		exit 1; \
	fi
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get configmap pod-report -n $(CLUSTER) -o jsonpath='{.data.report}' | jq .

# Lista todos os pod-reports de todos os clusters
check-all-reports:
	@echo "=== Pod Reports from all Spokes ==="
	@for ns in $$(KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get managedclusteraddon -A -o jsonpath='{.items[*].metadata.namespace}' 2>/dev/null); do \
		echo "--- $$ns ---"; \
		KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get configmap pod-report -n $$ns -o jsonpath='{.data.report}' 2>/dev/null | jq -c '{cluster: .clusterName, totalPods: .totalPods, timestamp: .timestamp}' 2>/dev/null || echo "No report"; \
	done

# Verifica AddOnPlacementScores
check-placement-scores:
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl get addonplacementscores -A

# Spoke Cluster Checks (uses dynamic SPOKE_KUBECONFIG based on SPOKE variable)

# Lista pods do agent no spoke
check-agent-pods:
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl get pods -n open-cluster-management-agent-addon

# Verifica logs do agent no spoke
check-agent-logs:
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl logs -l app=$(ADDON_NAME)-agent -n open-cluster-management-agent-addon --tail=50

# Lista ClusterClaims no spoke
check-claims:
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl get clusterclaims

# Lista namespaces no spoke
check-spoke-namespaces:
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl get ns

# Lista todos os recursos no spoke (debug)
check-spoke-all:
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl get all -A
