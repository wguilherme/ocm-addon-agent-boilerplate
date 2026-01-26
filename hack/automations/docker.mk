SHELL := /bin/bash

DOCKER_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(abspath $(DOCKER_MK_DIR)/../..)

include $(DOCKER_MK_DIR)variables.mk

.PHONY: build push kind-load tag

##@ Docker Image

build:
	@echo "Building Docker image: $(IMAGE)..."
	@docker build -t $(IMAGE) $(PROJECT_ROOT)
	@echo "Image built: $(IMAGE)"

push:
	@echo "Pushing Docker image: $(IMAGE)..."
	@docker push $(IMAGE)
	@echo "Image pushed: $(IMAGE)"

tag:
	@if [ -z "$(NEW_IMAGE)" ]; then \
		echo "Error: NEW_IMAGE variable not set. Usage: make docker-tag NEW_IMAGE=registry/image:tag"; \
		exit 1; \
	fi
	@docker tag $(IMAGE) $(NEW_IMAGE)
	@echo "Tagged $(IMAGE) as $(NEW_IMAGE)"

##@ Kind Integration

kind-load:
	@echo "Building and loading image into Kind clusters..."
	@docker build -t $(IMAGE) $(PROJECT_ROOT)
	@if kind get clusters 2>/dev/null | grep -q "^hub$$"; then \
		echo "Loading image into hub cluster..."; \
		kind load docker-image $(IMAGE) --name hub; \
	fi
	@for cluster in $$(kind get clusters 2>/dev/null | grep -v "^hub$$"); do \
		echo "Loading image into $$cluster cluster..."; \
		kind load docker-image $(IMAGE) --name $$cluster; \
	done
	@echo "Image loaded into all Kind clusters"

kind-load-hub:
	@echo "Loading image into hub cluster..."
	@kind load docker-image $(IMAGE) --name hub
	@echo "Image loaded into hub cluster"

kind-load-spoke:
	@if [ -z "$(SPOKE)" ]; then \
		echo "Error: SPOKE variable not set. Usage: make docker-kind-load-spoke SPOKE=spoke1"; \
		exit 1; \
	fi
	@echo "Loading image into $(SPOKE) cluster..."
	@kind load docker-image $(IMAGE) --name $(SPOKE)
	@echo "Image loaded into $(SPOKE) cluster"
