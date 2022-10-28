PROJECT = telemetry
PKG     = $(shell go list -m | head -1)
PKG_MAIN = cmd/telemetry/main.go
VERSION := $(or $(IMAGE_TAG),$(shell git describe --tags --first-parent --match "v*" 2> /dev/null || echo v0.0.0))
LOCAL_BIN_DIR := $(PWD)/bin

DOCKER_COMPOSE := docker-compose --project-name=$(PROJECT) --file=deployments/compose/compose.yml
DOCKER_SWARM := docker stack deploy --prune --with-registry-auth -c deployments/compose/compose.yml

## swaggo configuration
SWAG_VERSION := $(shell grep github.com/swaggo/swag go.mod | xargs echo | cut -d" " -f2)

## golangci configuration
GOLANGCI_CONFIG_URL   := https://gitlab.test.igdcs.com/finops/devops/cicd/runner/raw/master/.golangci.yml
GOLANGCI_LINT_VERSION := v1.49.0

DOCKER_IMAGE_NAME := telemetry:dev

.DEFAULT_GOAL := help

.PHONY: build docker golangci docs lint clean data test test-env env env-destroy run postgres postgres-destroy kafka kafka-producer kafka-consumer help

build: docs ## Build project
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X $(PKG)/config.AppVersion=$(VERSION)" -o $(PROJECT) $(PKG_MAIN)

docker: build
	tar -cf - deployments/docker/scratch.Dockerfile telemetry | docker build -t $(DOCKER_IMAGE_NAME) -f deployments/docker/scratch.Dockerfile -

bin/swag-$(SWAG_VERSION):
	@echo "> downloading swag@$(SWAG_VERSION)"
	@GOBIN=$(LOCAL_BIN_DIR) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	@mv $(LOCAL_BIN_DIR)/swag $(LOCAL_BIN_DIR)/swag-$(SWAG_VERSION)

.golangci.yml:
	@$(MAKE) golangci

golangci: ## Download .golangci.yml file
	@curl --insecure -o .golangci.yml -L'#' $(GOLANGCI_CONFIG_URL)

bin/golangci-lint-$(GOLANGCI_LINT_VERSION):
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCAL_BIN_DIR) $(GOLANGCI_LINT_VERSION)
	@mv $(LOCAL_BIN_DIR)/golangci-lint $(LOCAL_BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION)

docs: bin/swag-$(SWAG_VERSION) ## Generate swag documentation
	@$(LOCAL_BIN_DIR)/swag-$(SWAG_VERSION) init -g internal/http/server.go

lint: .golangci.yml bin/golangci-lint-$(GOLANGCI_LINT_VERSION) ## Lint Go files
	@$(LOCAL_BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION) --version
	@GOPATH="$(shell dirname $(PWD))" $(LOCAL_BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION) run ./...

clean: ## Remove binary
	@echo "> removing binary $(PROJECT)"
	@rm $(PROJECT) 2>/dev/null || true

data: ## Add testing data
	@go run ./dbtest/testdata/testdata.go

data-cleanup: ## Clean data
	@CLEANUP=1 go run ./dbtest/testdata/testdata.go

test: ## Run unit tests
	@go test -v -race -cover ./...

test-env: env data ## Run unit tests and integration tests
	go test -v -race -cover ./...
	@$(DOCKER_COMPOSE) down --volumes

env: ## Initializes a dev environment with dev dependencies
	@$(DOCKER_COMPOSE) up -d

env-ps: ## Check dev env
	@$(DOCKER_COMPOSE) ps

env-destroy: ## Stops the dependencies in the dev environment and destroys the data
	@$(DOCKER_COMPOSE) down --volumes

env-swarm: ## Initializes a dev envrionment in swarm
	@$(DOCKER_SWARM) $(PROJECT)

env-swarm-ps:
	docker stack ps $(PROJECT)

env-swarm-destroy:
	docker stack rm $(PROJECT)
	# wait for delete complete
	@until [[ -z "$(shell docker stack ps $(PROJECT) -q 2>/dev/null)" ]]; do sleep 1; done

# CONFIG_FILE=./configs/local.yml go run $(PKG_MAIN)
run: ## Run program
	OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
	OTEL_RESOURCE_ATTRIBUTES=service.name=telemetry,service.instance.id=1 \
	CONFIG_FILE=./configs/local.yml \
	go run $(PKG_MAIN)

run-docker: ## Run program in docker
	docker run -it --rm -p 8080:8080 --net $(PROJECT)_default -e OTEL_RESOURCE_ATTRIBUTES=service.name=telemetry $(DOCKER_IMAGE_NAME)

# OTEL_RESOURCE_ATTRIBUTES='"service.name={{slice .Service.Name 10}},service.instance.id={{.Task.ID}},host.id={{.Node.ID}},host.name={{.Node.Hostname}}"'
run-service: SCALE ?= 1
run-service: ## Run program as service
	docker service create -p 8080:8080 --network telemetry_default --name $(PROJECT)_$(PROJECT) \
	--replicas $(SCALE) \
	-e OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317 \
	-e OTEL_RESOURCE_ATTRIBUTES='service.name={{slice .Service.Name 10}},service.instance.id={{.Task.ID}},host.id={{.Node.ID}},host.name={{.Node.Hostname}}' \
	$(DOCKER_IMAGE_NAME)

rm-service: ## Remove program service
	docker service rm $(PROJECT)_$(PROJECT)

log-service: ## Log program service
	docker service logs $(PROJECT)_$(PROJECT) -f

scale-service: SCALE ?= 1
scale-service: ## Scale program SCALE=1
	docker service scale $(PROJECT)_$(PROJECT)=$(SCALE)

restart-collector: ## Restart collector service
	docker service update telemetry_otel-collector --force

postgres: ## Initialize a postgresql
	docker run -d --name postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust postgres:13.8-alpine

postgres-destroy: ## Destroy opened postgres
	docker container rm -f postgres

kafka-producer: ## Kafka console for producer
	@$(DOCKER_COMPOSE) exec kafka kafka-console-producer.sh --bootstrap-server localhost:9092 --topic default-topic

kafka-consumer: ## Kafka console for consumer
	@$(DOCKER_COMPOSE) exec kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic default-topic --from-beginning

kafka: ## Kafka container /bin/bash
	@$(DOCKER_COMPOSE) exec kafka /bin/bash

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
