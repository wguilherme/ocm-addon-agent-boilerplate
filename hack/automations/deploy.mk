SHELL := /bin/bash

DEPLOY_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(abspath $(DEPLOY_MK_DIR)/../..)
DEPLOY_DIR := $(PROJECT_ROOT)/deploy

include $(DEPLOY_MK_DIR)variables.mk

.PHONY: rbac controller full enable disable undeploy restart

##@ Deployment

rbac:
	@echo "Applying RBAC and ClusterManagementAddOn..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/serviceaccount.yaml
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/clusterrole.yaml
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/clusterrolebinding.yaml
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/managedclustersetbinding.yaml
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/placement.yaml
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/clustermanagementaddon.yaml
	@echo "RBAC and ClusterManagementAddOn applied"

controller:
	@echo "Deploying addon controller..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f $(DEPLOY_DIR)/deployment.yaml
	@echo "Waiting for controller to be ready..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl wait --for=condition=available deployment/$(ADDON_NAME)-controller -n open-cluster-management --timeout=120s
	@echo "Controller deployed and ready"

full: rbac controller
	@echo "Full deployment complete"

##@ Addon Management

enable:
	@if [ -z "$(CLUSTER)" ]; then \
		echo "Error: CLUSTER variable not set. Usage: make deploy-enable CLUSTER=spoke1"; \
		exit 1; \
	fi
	@echo "Enabling addon on cluster $(CLUSTER)..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl apply -f - <<EOF
	apiVersion: addon.open-cluster-management.io/v1alpha1
	kind: ManagedClusterAddOn
	metadata:
	  name: $(ADDON_NAME)
	  namespace: $(CLUSTER)
	spec:
	  installNamespace: open-cluster-management-agent-addon
	EOF
	@echo "Addon enabled on cluster $(CLUSTER)"

disable:
	@if [ -z "$(CLUSTER)" ]; then \
		echo "Error: CLUSTER variable not set. Usage: make deploy-disable CLUSTER=spoke1"; \
		exit 1; \
	fi
	@echo "Disabling addon on cluster $(CLUSTER)..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl delete managedclusteraddon $(ADDON_NAME) -n $(CLUSTER) --ignore-not-found
	@echo "Addon disabled on cluster $(CLUSTER)"

enable-all:
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		SPOKE_NAME="spoke$$i"; \
		echo "Enabling addon on $$SPOKE_NAME..."; \
		$(MAKE) -f $(DEPLOY_MK_DIR)deploy.mk enable CLUSTER=$$SPOKE_NAME; \
	done

disable-all:
	@for i in $$(seq 1 $(SPOKE_COUNT)); do \
		SPOKE_NAME="spoke$$i"; \
		echo "Disabling addon on $$SPOKE_NAME..."; \
		$(MAKE) -f $(DEPLOY_MK_DIR)deploy.mk disable CLUSTER=$$SPOKE_NAME; \
	done

##@ Cleanup

undeploy:
	@echo "Removing all addon resources from hub..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl delete -f $(DEPLOY_DIR)/ --ignore-not-found
	@echo "All addon resources removed"

##@ Operations

restart:
	@echo "Restarting addon controller..."
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl rollout restart deployment/$(ADDON_NAME)-controller -n open-cluster-management
	@KUBECONFIG=$$(eval echo $(HUB_KUBECONFIG)) kubectl wait --for=condition=available deployment/$(ADDON_NAME)-controller -n open-cluster-management --timeout=60s
	@echo "Controller restarted"

restart-agent:
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make deploy-restart-agent SPOKE=spoke1"; \
		exit 1; \
	fi
	@echo "Restarting agent on $(SPOKE)..."
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl rollout restart deployment/$(ADDON_NAME)-agent -n open-cluster-management-agent-addon 2>/dev/null || true
	@KUBECONFIG=$$(eval echo $(SPOKE_KUBECONFIG)) kubectl rollout status deployment/$(ADDON_NAME)-agent -n open-cluster-management-agent-addon --timeout=60s 2>/dev/null || true
	@echo "Agent restarted on $(SPOKE)"
