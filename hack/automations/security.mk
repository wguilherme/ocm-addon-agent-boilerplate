SHELL := /bin/bash

SECURITY_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(abspath $(SECURITY_MK_DIR)/../..)
REPORTS_DIR := $(PROJECT_ROOT)/reports

include $(SECURITY_MK_DIR)variables.mk

.PHONY: run gosec gitleaks trivy govulncheck hadolint

##@ Security Scanning

run: gosec gitleaks govulncheck
	@echo "Security checks completed. Reports saved to $(REPORTS_DIR)/"

gosec:
	@if ! command -v gosec &> /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@mkdir -p $(REPORTS_DIR)
	@echo "Running gosec..."
	@cd $(PROJECT_ROOT) && gosec -fmt=json -out=$(REPORTS_DIR)/gosec-report.json ./... 2>/dev/null || echo "Warning: Security issues detected (see report)"
	@echo "Gosec report: $(REPORTS_DIR)/gosec-report.json"

gitleaks:
	@mkdir -p $(REPORTS_DIR)
	@echo "Running gitleaks..."
	@docker run --rm \
		-v $(PROJECT_ROOT):/path \
		-v $(REPORTS_DIR):/reports \
		zricethezav/gitleaks:latest detect \
		--source="/path" \
		--report-path="/reports/gitleaks-report.json" \
		-v 2>/dev/null || echo "Warning: Secrets may have been detected (see report)"
	@echo "Gitleaks report: $(REPORTS_DIR)/gitleaks-report.json"

trivy:
	@mkdir -p $(REPORTS_DIR)
	@echo "Building image for scanning..."
	@docker build -t $(IMAGE) $(PROJECT_ROOT)
	@echo "Running trivy..."
	@docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(REPORTS_DIR):/reports \
		aquasec/trivy image \
		--exit-code 0 \
		--ignore-unfixed \
		--format json \
		--output /reports/trivy-report.json \
		$(IMAGE) 2>/dev/null || echo "Warning: Vulnerabilities detected (see report)"
	@echo "Trivy report: $(REPORTS_DIR)/trivy-report.json"

govulncheck:
	@if ! command -v govulncheck &> /dev/null; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@mkdir -p $(REPORTS_DIR)
	@echo "Running govulncheck..."
	@cd $(PROJECT_ROOT) && govulncheck -json ./... > $(REPORTS_DIR)/govulncheck-report.json 2>/dev/null || echo "Warning: Vulnerabilities detected (see report)"
	@echo "Govulncheck report: $(REPORTS_DIR)/govulncheck-report.json"

hadolint:
	@echo "Running hadolint on Dockerfile..."
	@docker run --rm -i \
		-e HADOLINT_FAILURE_THRESHOLD=$(or $(THRESHOLD),warning) \
		hadolint/hadolint < $(PROJECT_ROOT)/Dockerfile || echo "Warning: Dockerfile issues detected"
