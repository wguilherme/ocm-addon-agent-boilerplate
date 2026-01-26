SHELL := /bin/bash

CORE_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
include $(CORE_MK_DIR)variables.mk

.PHONY: all setup clean prepare kind-% ocm-% deploy-% test-% security-% docker-%

##@ Main Workflows

all: kind-cluster ocm-init-hub ocm-join-all-spokes deploy-rbac
	@echo ""
	@echo "=== Setup Complete ==="
	@echo "Hub and $(SPOKE_COUNT) spoke(s) ready with OCM"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run the controller: make run"
	@echo "  2. Enable addon: make deploy-enable CLUSTER=spoke1"
	@echo "  3. Check reports: make test-check-report CLUSTER=spoke1"

setup: setup-install-lefthook
	@echo "Development environment setup complete"

prepare: kind-cluster kind-get-all-kubeconfigs ocm-init-hub ocm-join-all-spokes docker-kind-load deploy-full
	@echo ""
	@echo "=== Full Environment Ready ==="
	@echo "Kubeconfigs exported to: $(KUBECONFIG_KIND_EXTRACT_CONFIG_PATH)/"
	@echo ""
	@echo "To check addon reports:"
	@echo "  make test-check-all-reports"

clean: kind-clean
	@rm -f join-command.txt
	@echo "Environment cleaned"

##@ Pattern Routers (delegate to specific .mk files)

kind-%:
	@$(MAKE) -f $(CORE_MK_DIR)kind.mk $(patsubst kind-%,%,$@)

ocm-%:
	@$(MAKE) -f $(CORE_MK_DIR)ocm.mk $(patsubst ocm-%,%,$@)

deploy-%:
	@$(MAKE) -f $(CORE_MK_DIR)deploy.mk $(patsubst deploy-%,%,$@)

test-%:
	@$(MAKE) -f $(CORE_MK_DIR)test.mk $(patsubst test-%,%,$@)

security-%:
	@$(MAKE) -f $(CORE_MK_DIR)security.mk $(patsubst security-%,%,$@)

docker-%:
	@$(MAKE) -f $(CORE_MK_DIR)docker.mk $(patsubst docker-%,%,$@)

setup-%:
	@$(MAKE) -f $(CORE_MK_DIR)setup.mk $(patsubst setup-%,%,$@)
