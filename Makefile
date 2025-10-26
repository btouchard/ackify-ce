# SPDX-License-Identifier: AGPL-3.0-or-later
# Makefile for ackify-ce project

.PHONY: build build-frontend build-backend build-all test test-unit test-integration test-short coverage lint fmt vet clean help dev dev-frontend dev-backend migrate-up migrate-down docker-rebuild

# Variables
BINARY_NAME=ackify-ce
BUILD_DIR=./backend/cmd/community
MIGRATE_DIR=./backend/cmd/migrate
COVERAGE_DIR=coverage
WEBAPP_DIR=./webapp

# Default target
help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: build-all ## Build the complete application (frontend + backend)

build-frontend: ## Build the Vue.js frontend
	@echo "Building frontend..."
	cd $(WEBAPP_DIR) && npm install && npm run build

build-backend: ## Build the Go backend
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(BUILD_DIR)

build-all: build-frontend build-backend ## Build frontend and backend

# Test targets
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests with race detection..."
	CGO_ENABLED=1 go test -short -race -v ./backend/internal/... ./backend/pkg/... ./backend/cmd/...

test-integration: ## Run integration tests (requires PostgreSQL - migrations are applied automatically)
	@echo "Running integration tests with race detection..."
	@echo "Note: Migrations are applied automatically by test setup"
	@export INTEGRATION_TESTS=1; \
	export ACKIFY_DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"; \
	CGO_ENABLED=1 go test -v -race -tags=integration ./backend/internal/infrastructure/database/...

test-integration-setup: ## Setup test database for integration tests (migrations applied by tests)
	@echo "Setting up test database..."
	@psql "postgres://postgres:testpassword@localhost:5432/postgres?sslmode=disable" -c "DROP DATABASE IF EXISTS ackify_test;" || true
	@psql "postgres://postgres:testpassword@localhost:5432/postgres?sslmode=disable" -c "CREATE DATABASE ackify_test;"
	@echo "Test database ready! Migrations will be applied automatically when tests run."

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
	@export ACKIFY_DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"; \
	export INTEGRATION_TESTS=1; \
	CGO_ENABLED=1 go test -v -race -tags=integration -coverprofile=$(COVERAGE_DIR)/coverage-integration.out ./backend/internal/infrastructure/database/...
	go tool cover -html=$(COVERAGE_DIR)/coverage-integration.out -o $(COVERAGE_DIR)/coverage-integration.html
	@echo "Integration coverage report generated: $(COVERAGE_DIR)/coverage-integration.html"

coverage-all: ## Generate full coverage report (unit + integration merged)
	@echo "Generating full coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@echo "Step 1/3: Running unit tests with coverage..."
	@CGO_ENABLED=1 go test -short -race -coverprofile=$(COVERAGE_DIR)/coverage-unit.out ./backend/internal/... ./backend/pkg/... ./backend/cmd/...
	@echo "Step 2/3: Running integration tests with coverage..."
	@export ACKIFY_DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"; \
	export INTEGRATION_TESTS=1; \
	CGO_ENABLED=1 go test -v -race -tags=integration -coverprofile=$(COVERAGE_DIR)/coverage-integration.out ./backend/internal/infrastructure/database/...
	@echo "Step 3/3: Merging coverage files..."
	@echo "mode: atomic" > $(COVERAGE_DIR)/coverage-all.out
	@grep -h -v "^mode:" $(COVERAGE_DIR)/coverage-unit.out $(COVERAGE_DIR)/coverage-integration.out >> $(COVERAGE_DIR)/coverage-all.out || true
	@go tool cover -html=$(COVERAGE_DIR)/coverage-all.out -o $(COVERAGE_DIR)/coverage-all.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage-all.out
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
dev: dev-backend ## Start development server (backend only - frontend served by backend)

dev-frontend: ## Start frontend development server (Vite hot reload)
	@echo "Starting frontend dev server..."
	cd $(WEBAPP_DIR) && npm run dev

dev-backend: ## Run backend in development mode
	@echo "Starting backend..."
	go run $(BUILD_DIR)

clean: ## Clean build artifacts and test coverage
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(COVERAGE_DIR)
	rm -rf $(WEBAPP_DIR)/dist
	rm -rf $(WEBAPP_DIR)/node_modules
	go clean ./...

deps: ## Download and tidy dependencies (Go + npm)
	@echo "Downloading Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing frontend dependencies..."
	cd $(WEBAPP_DIR) && npm install

migrate-up: ## Apply database migrations
	@echo "Applying database migrations..."
	go run $(MIGRATE_DIR) up

migrate-down: ## Rollback last database migration
	@echo "Rolling back last migration..."
	go run $(MIGRATE_DIR) down

# Mock generation (none at the moment)
generate-mocks: ## No exported interfaces to mock (skipped)
	@echo "Skipping mock generation: no exported interfaces to mock."

# Docker targets
docker-build: ## Build Docker image
	docker build -t ackify-ce:latest .

docker-rebuild: ## Rebuild and restart Docker containers (as per CLAUDE.md)
	@echo "Rebuilding and restarting Docker containers..."
	docker compose -f compose.local.yml up -d --force-recreate ackify-ce --build

docker-up: ## Start Docker containers
	docker compose -f compose.local.yml up -d

docker-down: ## Stop Docker containers
	docker compose -f compose.local.yml down

docker-logs: ## View Docker logs
	docker compose -f compose.local.yml logs -f ackify-ce

docker-test: ## Run tests in Docker environment
	docker compose -f compose.local.yml up -d ackify-db
	@sleep 5
	$(MAKE) test
	docker compose -f compose.local.yml down

# CI targets
ci: deps lint test coverage ## Run all CI checks

# Install dev tools
dev-tools: ## Install development tools
	@echo "Installing development tools..."
	go install go.uber.org/mock/mockgen@latest
