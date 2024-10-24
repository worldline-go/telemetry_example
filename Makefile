PROJECT = telemetry
PKG     = $(shell go list -m | head -1)
PKG_MAIN = cmd/$(PROJECT)/main.go
VERSION := $(or $(IMAGE_TAG),$(shell git describe --tags --first-parent --match "v*" 2> /dev/null || echo v0.0.0))
LOCAL_BIN_DIR := $(PWD)/bin

DOCKER_IMAGE_NAME := telemetry:test

.DEFAULT_GOAL := help

.PHONY: build
build: ## Build project
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X $(PKG)/config.AppVersion=$(VERSION)" -o dist/$(PROJECT) $(PKG_MAIN)

.PHONY: build-container
build-container: build ## Build project and docker image
	tar -cf - deployments/docker/scratch.Dockerfile dist/telemetry migrations | docker build -t $(DOCKER_IMAGE_NAME) -f deployments/docker/scratch.Dockerfile -

.PHONY: docs
docs: ## Generate swag documentation
	@swag init -g internal/server/server.go

.PHONY: install
install: ## Install tools for development
	go install github.com/swaggo/swag/cmd/swag@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/bin v1.61.0

.PHONY: lint
lint: ## Lint Go files
	@GOPATH="$(shell dirname $(PWD))" golangci-lint run ./...

.PHONY: clean
clean: ## Remove binary
	@echo "> removing binary $(PROJECT)"
	@rm $(PROJECT) 2>/dev/null || true

.PHONY: data
data: ## Add testing data
	@go run ./dbtest/testdata/testdata.go

.PHONY: data-cleanup
data-cleanup: ## Clean data
	@CLEANUP=1 go run ./dbtest/testdata/testdata.go

.PHONY: test
test: ## Run unit tests
	@go test -v -race -cover ./...

.PHONY: test-env
test-env: env data ## Run unit tests and integration tests
	go test -v -race -cover ./...
	@$(DOCKER_COMPOSE) down --volumes

.PHONY: env
env: ## Initializes a dev environment with dev dependencies
	docker compose --project-name=telemetry --file=deployments/compose/compose.yml up -d --remove-orphans

.PHONY: env-services
env-services: ## Initializes a services environment
	docker compose --project-name=telemetry-services --file=deployments/services/compose.yml up -d --remove-orphans

.PHONY: env-destroy
env-destroy: ## Stops the dependencies in the dev environment and destroys the data
	docker compose --project-name=telemetry down --volumes

.PHONY: env-services-destroy
env-services-destroy: ## Stops the dependencies in the services environment and destroys the data
	docker compose --project-name=telemetry-services down --volumes

# CONFIG_FILE=./configs/local.yml go run $(PKG_MAIN)
.PHONY: run
run: ## Run program
	OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
	OTEL_RESOURCE_ATTRIBUTES=service.name=telemetry,service.instance.id=1 \
	CONFIG_FILE=./configs/local.yml \
	go run $(PKG_MAIN)

.PHONY: run-without
run-without: ## Run program without telemetry
	CONFIG_FILE=./configs/local.yml \
	go run $(PKG_MAIN)

.PHONY: run-docker
run-docker: ## Run program in docker
	docker run -it --rm -p 8080:8080 --net $(PROJECT)_default -e OTEL_RESOURCE_ATTRIBUTES=service.name=telemetry $(DOCKER_IMAGE_NAME)

# OTEL_RESOURCE_ATTRIBUTES='"service.name={{slice .Service.Name 10}},service.instance.id={{.Task.ID}},host.id={{.Node.ID}},host.name={{.Node.Hostname}}"'
.PHONY: run-service
run-service: SCALE ?= 1
run-service: ## Run program as service
	docker service create -p 8080:8080 --network telemetry_default --name $(PROJECT)_$(PROJECT) \
	--replicas $(SCALE) \
	-e OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317 \
	-e OTEL_RESOURCE_ATTRIBUTES='service.name={{slice .Service.Name 10}},service.instance.id={{.Task.ID}},host.id={{.Node.ID}},host.name={{.Node.Hostname}}' \
	$(DOCKER_IMAGE_NAME)

.PHONY: rm-service
rm-service: ## Remove program service
	docker service rm $(PROJECT)_$(PROJECT)

.PHONY: log-service
log-service: ## Log program service
	docker service logs $(PROJECT)_$(PROJECT) -f

.PHONY: scale-service
scale-service: SCALE ?= 1
scale-service: ## Scale program SCALE=1
	docker service scale $(PROJECT)_$(PROJECT)=$(SCALE)

.PHONY: restart-collector
restart-collector: ## Restart collector service
	docker service update telemetry_otel-collector --force

.PHONY: postgres
postgres: ## Initialize a postgresql
	docker run -d --name postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust postgres:13.8-alpine

.PHONY: postgres-destroy
postgres-destroy: ## Destroy opened postgres
	docker container rm -f postgres

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
