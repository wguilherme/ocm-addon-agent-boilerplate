SHELL := /bin/bash

OCM_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
include $(OCM_MK_DIR)variables.mk

.PHONY: install-clusteradm init-hub join-spoke accept-spoke status restart

##@ OCM Tools

install-clusteradm:
	@command -v clusteradm &>/dev/null || { \
		echo "Installing clusteradm..."; \
		curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash; \
	}

##@ Hub Initialization

init-hub: install-clusteradm
	@echo "Initializing OCM hub..."
	@clusteradm init \
		--use-bootstrap-token \
		--feature-gates=ManifestWorkReplicaSet=true,ManagedClusterAutoApproval=true \
		--wait=true \
		--context kind-hub | grep "clusteradm join" > join-command.txt
	@echo "Hub initialized. Join command saved to join-command.txt"

##@ Spoke Management

join-spoke: install-clusteradm
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make ocm-join-spoke SPOKE=spoke1"; \
		exit 1; \
	fi
	@[ -f ./join-command.txt ] || { echo "Error: join-command.txt not found. Run 'make ocm-init-hub' first."; exit 1; }
	@echo "Joining spoke $(SPOKE) to hub..."
	@if kubectl --context kind-hub get managedcluster $(SPOKE) -o jsonpath='{.spec.hubAcceptsClient}' 2>/dev/null | grep -q "true"; then \
		echo "Spoke $(SPOKE) already joined and accepted"; \
		exit 0; \
	fi
	@HUB_TOKEN=$$(cat join-command.txt | awk -F'--hub-token ' '{print $$2}' | awk '{print $$1}' | tr -d '\n'); \
	HUB_APISERVER=$$(cat join-command.txt | awk -F'--hub-apiserver ' '{print $$2}' | awk '{print $$1}' | tr -d '\n'); \
	clusteradm join \
		--hub-token $$HUB_TOKEN \
		--hub-apiserver $$HUB_APISERVER \
		--cluster-name $(SPOKE) \
		--feature-gates=AddonManagement=true,ClusterClaim=true \
		--force-internal-endpoint-lookup \
		--context kind-$(SPOKE); \
	echo "Waiting for spoke $(SPOKE) to appear on hub..."; \
	until kubectl --context kind-hub get managedcluster $(SPOKE) >/dev/null 2>&1; do sleep 2; done; \
	echo "Accepting spoke $(SPOKE)..."; \
	clusteradm accept --clusters $(SPOKE) --context kind-hub
	@echo "Spoke $(SPOKE) joined and accepted"

join-all-spokes:
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		SPOKE_NAME="spoke$$i"; \
		echo "Joining spoke $$SPOKE_NAME..."; \
		$(MAKE) -f $(OCM_MK_DIR)ocm.mk join-spoke SPOKE=$$SPOKE_NAME; \
	done

accept-spoke:
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make ocm-accept-spoke SPOKE=spoke1"; \
		exit 1; \
	fi
	@echo "Accepting spoke $(SPOKE)..."
	@clusteradm accept --clusters $(SPOKE) --context kind-hub
	@echo "Spoke $(SPOKE) accepted"

##@ Status & Troubleshooting

status:
	@echo "=== Managed Clusters ==="
	@kubectl --context kind-hub get managedclusters
	@echo ""
	@echo "=== Managed Cluster Addons ==="
	@kubectl --context kind-hub get managedclusteraddons -A
	@echo ""
	@echo "=== Cluster Manager Status ==="
	@kubectl --context kind-hub get pods -n open-cluster-management-hub

restart:
	@echo "Restarting OCM components..."
	@kubectl --context kind-hub -n open-cluster-management rollout restart deployment cluster-manager 2>/dev/null || true
	@kubectl --context kind-hub -n open-cluster-management-hub rollout restart deployment cluster-manager-work-webhook 2>/dev/null || true
	@kubectl --context kind-hub -n open-cluster-management-hub rollout restart deployment cluster-manager-work-controller 2>/dev/null || true
	@kubectl --context kind-hub -n open-cluster-management-hub rollout restart deployment cluster-manager-registration-webhook 2>/dev/null || true
	@kubectl --context kind-hub -n open-cluster-management-hub rollout restart deployment cluster-manager-registration-controller 2>/dev/null || true
	@kubectl --context kind-hub -n open-cluster-management-hub rollout restart deployment cluster-manager-placement-controller 2>/dev/null || true
	@kubectl --context kind-hub -n open-cluster-management-hub rollout restart deployment cluster-manager-addon-manager-controller 2>/dev/null || true
	@echo "OCM components restarted"
