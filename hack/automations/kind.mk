SHELL := /bin/bash

KIND_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(abspath $(KIND_MK_DIR)/../..)
-include $(PROJECT_ROOT)/.env
export

include $(KIND_MK_DIR)variables.mk

# Kind config files
KIND_HUB_CONFIG   := $(KIND_MK_DIR)manifests/kind-hub.yaml
KIND_SPOKE_CONFIG := $(KIND_MK_DIR)manifests/kind-spoke.yaml

.PHONY: tools install-kind install-kubectl
.PHONY: registry cluster hub-cluster spoke-clusters spoke-cluster
.PHONY: clean clean-hub clean-spokes clean-spoke
.PHONY: get-kubeconfig get-all-kubeconfigs

##@ Kind Tools

tools: install-kind install-kubectl

install-kind:
	@command -v kind &>/dev/null || { \
		echo "Installing kind $(KIND_VERSION)..."; \
		curl -Lo ./kind https://kind.sigs.k8s.io/dl/$(KIND_VERSION)/kind-$$(uname -s | tr '[:upper:]' '[:lower:]')-$$(uname -m | sed 's/x86_64/amd64/'); \
		chmod +x ./kind; \
		sudo mv ./kind /usr/local/bin/kind; \
	}

install-kubectl:
	@command -v kubectl &>/dev/null || { \
		echo "Installing kubectl $(KUBECTL_VERSION)..."; \
		curl -LO "https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$$(uname -s | tr '[:upper:]' '[:lower:]')/amd64/kubectl"; \
		chmod +x ./kubectl; \
		sudo mv ./kubectl /usr/local/bin/kubectl; \
	}

##@ Registry

registry:
	@if [ "$$(docker inspect -f '{{.State.Running}}' kind-registry 2>/dev/null || echo 'false')" != 'true' ]; then \
		docker rm -f kind-registry 2>/dev/null || true; \
		docker run -d --restart=always -p "127.0.0.1:5000:5000" --network bridge --name "kind-registry" registry:2; \
	fi
	@docker network create kind 2>/dev/null || true
	@docker network connect kind kind-registry 2>/dev/null || true
	@echo "Registry running at localhost:5000"

##@ Clusters

cluster: hub-cluster spoke-clusters

hub-cluster: registry tools
	@if ! kind get clusters 2>/dev/null | grep -q "^hub$$"; then \
		echo "Creating hub cluster..."; \
		kind create cluster --config $(KIND_HUB_CONFIG) --name hub; \
	elif [ "$$(docker inspect -f '{{.State.Running}}' hub-control-plane 2>/dev/null)" != "true" ]; then \
		echo "Starting hub cluster..."; \
		docker start hub-control-plane; \
	fi
	@kubectl --context kind-hub wait --for=condition=Ready nodes --all --timeout=120s
	@echo "Hub cluster ready"

spoke-clusters: registry tools
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		SPOKE_NAME="spoke$$i"; \
		echo "Creating spoke cluster: $$SPOKE_NAME"; \
		$(MAKE) -f $(KIND_MK_DIR)kind.mk spoke-cluster SPOKE=$$SPOKE_NAME; \
	done

spoke-cluster: registry tools
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make kind-spoke-cluster SPOKE=spoke1"; \
		exit 1; \
	fi
	@if ! kind get clusters 2>/dev/null | grep -q "^$(SPOKE)$$"; then \
		echo "Creating spoke cluster: $(SPOKE)..."; \
		kind create cluster --config $(KIND_SPOKE_CONFIG) --name $(SPOKE); \
	elif [ "$$(docker inspect -f '{{.State.Running}}' $(SPOKE)-control-plane 2>/dev/null)" != "true" ]; then \
		echo "Starting spoke cluster: $(SPOKE)..."; \
		docker start $(SPOKE)-control-plane; \
		kubectl --context kind-$(SPOKE) wait --for=condition=Ready nodes --all --timeout=60s 2>/dev/null || true; \
	fi
	@echo "Spoke cluster $(SPOKE) ready"

##@ Cleanup

clean: clean-hub clean-spokes
	@rm -f join-command.txt
	@docker rm -f kind-registry 2>/dev/null || true
	@echo "All clusters and registry cleaned"

clean-hub:
	@kind delete cluster --name hub 2>/dev/null || true
	@echo "Hub cluster deleted"

clean-spokes:
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		SPOKE_NAME="spoke$$i"; \
		echo "Deleting spoke cluster: $$SPOKE_NAME"; \
		kind delete cluster --name $$SPOKE_NAME 2>/dev/null || true; \
	done

clean-spoke:
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make kind-clean-spoke SPOKE=spoke1"; \
		exit 1; \
	fi
	@kind delete cluster --name $(SPOKE) 2>/dev/null || true
	@echo "Spoke cluster $(SPOKE) deleted"

##@ Kubeconfig

get-kubeconfig:
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make kind-get-kubeconfig SPOKE=spoke1"; \
		exit 1; \
	fi
	@mkdir -p $(KUBECONFIG_PATH_EXPANDED)
	@kind get kubeconfig --name $(SPOKE) > $(KUBECONFIG_PATH_EXPANDED)/config.$(SPOKE)
	@echo "Kubeconfig exported to: $(KUBECONFIG_KIND_EXTRACT_CONFIG_PATH)/config.$(SPOKE)"

get-all-kubeconfigs:
	@mkdir -p $(KUBECONFIG_PATH_EXPANDED)
	@kind get kubeconfig --name hub > $(KUBECONFIG_PATH_EXPANDED)/config.hub
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		SPOKE_NAME="spoke$$i"; \
		kind get kubeconfig --name $$SPOKE_NAME > $(KUBECONFIG_PATH_EXPANDED)/config.$$SPOKE_NAME; \
	done
	@echo "Kubeconfigs exported to: $(KUBECONFIG_KIND_EXTRACT_CONFIG_PATH)/"
	@echo "  - Hub: config.hub"
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		echo "  - Spoke$$i: config.spoke$$i"; \
	done
