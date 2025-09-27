# SPDX-License-Identifier: AGPL-3.0-or-later
# Makefile for ackify-ce project

.PHONY: build test test-unit test-integration test-short coverage lint fmt vet clean help

# Variables
BINARY_NAME=ackify-ce
BUILD_DIR=./cmd/community
COVERAGE_DIR=coverage

# Default target
help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(BUILD_DIR)

# Test targets
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests with race detection..."
	CGO_ENABLED=1 go test -short -race -v ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests (requires PostgreSQL)
	@echo "Running integration tests with race detection..."
	@if [ -z "$(DB_DSN)" ]; then \
		export DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"; \
	fi; \
	export INTEGRATION_TESTS=true; \
	CGO_ENABLED=1 go test -v -race -tags=integration ./internal/infrastructure/database/...

test-integration-setup: ## Setup test database for integration tests
	@echo "Setting up test database..."
	@psql "postgres://postgres:testpassword@localhost:5432/postgres?sslmode=disable" -c "DROP DATABASE IF EXISTS ackify_test;" || true
	@psql "postgres://postgres:testpassword@localhost:5432/postgres?sslmode=disable" -c "CREATE DATABASE ackify_test;"
	@echo "Test database ready!"

test-short: ## Run only quick tests
	@echo "Running short tests..."
	go test -short ./...


# Coverage targets
coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

coverage-integration: ## Generate integration test coverage report
	@echo "Generating integration coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@export DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"; \
	export INTEGRATION_TESTS=true; \
	go test -v -race -tags=integration -coverprofile=$(COVERAGE_DIR)/coverage-integration.out ./internal/infrastructure/database/...
	go tool cover -html=$(COVERAGE_DIR)/coverage-integration.out -o $(COVERAGE_DIR)/coverage-integration.html
	@echo "Integration coverage report generated: $(COVERAGE_DIR)/coverage-integration.html"

coverage-all: ## Generate full coverage report (unit + integration)
	@echo "Generating full coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@export DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"; \
	export INTEGRATION_TESTS=true; \
	go test -v -race -tags=integration -coverprofile=$(COVERAGE_DIR)/coverage-all.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage-all.out -o $(COVERAGE_DIR)/coverage-all.html
	go tool cover -func=$(COVERAGE_DIR)/coverage-all.out
	@echo "Full coverage report generated: $(COVERAGE_DIR)/coverage-all.html"

coverage-func: ## Show function-level coverage
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -func=$(COVERAGE_DIR)/coverage.out

# Code quality targets
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: fmt vet ## Run basic linting tools

lint-extra: ## Run staticcheck if available (installs if missing)
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; }
	staticcheck ./...

# Development targets
clean: ## Clean build artifacts and test coverage
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(COVERAGE_DIR)
	go clean ./...

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Mock generation (none at the moment)
generate-mocks: ## No exported interfaces to mock (skipped)
	@echo "Skipping mock generation: no exported interfaces to mock."

# Docker targets
docker-build: ## Build Docker image
	docker build -t ackify-ce:latest .

docker-test: ## Run tests in Docker environment
	docker compose -f docker-compose.local.yml up -d ackify-db
	@sleep 5
	$(MAKE) test
	docker compose -f docker-compose.local.yml down

# CI targets
ci: deps lint test coverage ## Run all CI checks

# Install dev tools
dev-tools: ## Install development tools
	@echo "Installing development tools..."
	go install go.uber.org/mock/mockgen@latest
