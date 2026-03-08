export GO111MODULE=on

#===================#
#== Env Variables ==#
#===================#
DOCKER_COMPOSE_FILE ?= docker-compose.yml
RUN_IN_DOCKER       ?= docker compose exec builder
BINARY_NAME         ?= app

help: ## Show this help
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##/:/'`); \
	printf "%-25s %s\n" "target" "help" ; \
	printf "%-25s %s\n" "------" "----" ; \
	for help_line in $${help_lines[@]}; do \
		IFS=$$':' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf '\033[36m%-25s\033[0m %s\n' $$help_command $$help_info; \
	done

#===============#
#== App Build ==#
#===============#

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build -o ./bin/$(BINARY_NAME) ./cmd/main.go

build-docker: ## Build the Docker image
	docker build -t $(BINARY_NAME):latest . --no-cache

#===============#
#=== App Run ===#
#===============#

run-http: build ## Run HTTP server
	./bin/$(BINARY_NAME) http

run-grpc: build ## Run gRPC server
	./bin/$(BINARY_NAME) grpc

run-consumer: build ## Run message consumer
	./bin/$(BINARY_NAME) consumer

#=======================#
#== ENVIRONMENT SETUP ==#
#=======================#

env: ## Copy .env.sample to .env if it doesn't exist
ifeq (,$(wildcard .env))
	cp .env.sample .env
endif

clean: ## Remove built binaries
	rm -rf bin/*

#======================#
#== DATABASE / MIGRATE ==#
#======================#

migrate-up: ## Run all pending migrations
	docker compose up migrate

migrate-create: ## Create a new migration file: make migrate-create name=<migration-name>
	docker compose run --rm migrate create -ext sql -dir /migrations $(name)

#====================#
#== QUALITY CHECKS ==#
#====================#

lint: ## Run golangci-lint
	docker run -t --rm -v ${PWD}:/app -v $$(go env GOMODCACHE):/go/pkg/mod \
		-w /app golangci/golangci-lint:v2.7.0 golangci-lint run -v

test: ## Run unit tests
	@echo "Running unit tests..."
	go test -tags unit -shuffle=on \
		$$(go list ./... | grep -v mock | grep -v generated | tr '\n' ' ') \
		-coverpkg=./... -coverprofile coverage.out

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -tags integration -shuffle=on ./...

test-load: ## Run a k6 load test: make test-load name=http_post_create_example
	./scripts/k6.sh $(name)

test-load-list: ## List available k6 load tests
	./scripts/k6.sh

fmt: ## Format code
	gci write -s standard -s default . --skip-generated --skip-vendor && \
	gofumpt -l -w .

generate: ## Run go generate across all packages
	go generate ./...

#==========================#
#== DOCKER INFRASTRUCTURE ==#
#==========================#

docker-start: ## Start Docker environment
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d --build --remove-orphans

docker-stop: ## Stop Docker environment
	docker compose -f $(DOCKER_COMPOSE_FILE) stop

docker-clean: docker-stop ## Remove Docker containers and volumes
	docker compose -f $(DOCKER_COMPOSE_FILE) rm -v -f

docker-restart: docker-stop docker-start ## Restart Docker environment

#==========================#
#== OPENAPI CODEGEN       ==#
#==========================#

generate-contracts: ## Regenerate HTTP API types from OpenAPI spec
	@mkdir -p internal/app/transport/http/api-contract
	docker run --rm -v "`pwd`:/app" -w /app hidori/oapi-codegen:latest \
		--config docs/openapi/oapi-codegen.yaml docs/openapi/openapi.yaml

.PHONY: help build build-docker run-http run-grpc run-consumer env clean \
        migrate-up migrate-create lint test test-integration test-load test-load-list fmt generate \
        docker-start docker-stop docker-clean docker-restart generate-contracts
