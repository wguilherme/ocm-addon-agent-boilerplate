SHELL := /bin/bash

.PHONY: install-lefthook install-tools

##@ Development Setup

install-lefthook:
	@if ! command -v lefthook &> /dev/null; then \
		echo "Installing lefthook..."; \
		go install github.com/evilmartians/lefthook@latest; \
	fi
	@lefthook install
	@echo "Lefthook installed and git hooks configured"

install-tools:
	@echo "Installing development tools..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo "Development tools installed"
