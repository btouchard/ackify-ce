#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Complete test suite runner with coverage reporting for Ackify CE

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Ackify CE - Complete Test Suite Runner           ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "webapp" ]; then
    echo -e "${RED}❌ Error: Please run this script from the project root directory${NC}"
    exit 1
fi

# Variables
PROJECT_ROOT=$(pwd)
WEBAPP_DIR="$PROJECT_ROOT/webapp"
COVERAGE_DIR="$PROJECT_ROOT/.coverage-report"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Track failures
FAILED=0
INTEGRATION_SKIPPED=0
E2E_SKIPPED=0

# Cleanup function
cleanup_integration() {
    if [ "$INTEGRATION_STARTED" = "1" ]; then
        echo ""
        echo -e "${YELLOW}🧹 Cleaning up integration test environment...${NC}"
        docker compose -f compose.test.yml down -v --remove-orphans > /dev/null 2>&1 || true
        echo -e "${GREEN}✓ Integration test environment cleaned up${NC}"
    fi
}

cleanup_e2e() {
    if [ "$E2E_STARTED" = "1" ]; then
        echo ""
        echo -e "${YELLOW}🧹 Cleaning up E2E test environment...${NC}"
        docker compose -f compose.e2e.yml down -v --remove-orphans > /dev/null 2>&1 || true
        echo -e "${GREEN}✓ E2E test environment cleaned up${NC}"
    fi
}

# Trap to ensure cleanup on exit
trap cleanup_integration EXIT

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

# ==============================================================================
# Phase 1: Backend Tests
# ==============================================================================
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Phase 1/3: Backend Tests (Go)${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$PROJECT_ROOT"

echo -e "${YELLOW}📦 Running go fmt check...${NC}"
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo -e "${RED}❌ Code formatting issues found:${NC}"
    gofmt -s -l .
    FAILED=$((FAILED + 1))
else
    echo -e "${GREEN}✓ Code formatting OK${NC}"
fi
echo ""

echo -e "${YELLOW}🔍 Running go vet...${NC}"
if go vet ./...; then
    echo -e "${GREEN}✓ go vet passed${NC}"
else
    echo -e "${RED}❌ go vet failed${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

echo -e "${YELLOW}🧪 Running unit tests...${NC}"
if go test -v -short ./...; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}❌ Unit tests failed${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

# Generate unit coverage
echo -e "${YELLOW}📊 Generating unit test coverage...${NC}"
go test -short -covermode=atomic -coverprofile="$COVERAGE_DIR/backend-unit.out" ./... 2>&1 | grep -v "no test files" || true
echo -e "${GREEN}✓ Unit coverage generated${NC}"
echo ""

# Integration tests with Docker Compose
echo -e "${YELLOW}🔗 Running integration tests...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️  Docker not available, skipping integration tests${NC}"
    INTEGRATION_SKIPPED=1
else
    echo -e "${YELLOW}🐳 Starting PostgreSQL + MailHog (compose.test.yml)...${NC}"

    # Clean up previous containers
    docker compose -f "$PROJECT_ROOT/compose.test.yml" down -v --remove-orphans > /dev/null 2>&1 || true

    # Start services
    if docker compose -f "$PROJECT_ROOT/compose.test.yml" up -d --remove-orphans; then
        INTEGRATION_STARTED=1
        echo -e "${GREEN}✓ Services started${NC}"

        # Wait for PostgreSQL to be ready
        echo -e "${YELLOW}⏳ Waiting for PostgreSQL to be ready...${NC}"
        sleep 5

        MAX_RETRIES=30
        RETRY_COUNT=0
        while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
            if docker exec ackify-db-test pg_isready -U postgres -d ackify_test > /dev/null 2>&1; then
                echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
                break
            fi
            RETRY_COUNT=$((RETRY_COUNT + 1))
            echo "   Retry $RETRY_COUNT/$MAX_RETRIES..."
            sleep 2
        done

        if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
            echo -e "${RED}❌ PostgreSQL failed to start${NC}"
            FAILED=$((FAILED + 1))
            INTEGRATION_SKIPPED=1
        else
            # Run migrations
            echo -e "${YELLOW}📝 Running database migrations...${NC}"
            export ACKIFY_DB_DSN="postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"
            export ACKIFY_APP_PASSWORD="ackifytestpassword"
            cd "$PROJECT_ROOT"
            if go run ./backend/cmd/migrate/main.go -migrations-path file://backend/migrations up; then
                echo -e "${GREEN}✓ Migrations applied${NC}"

                # Run integration tests
                export INTEGRATION_TESTS=1
                if go test -v -race -tags=integration -p 1 -count=1 ./backend/internal/infrastructure/database/... ./backend/internal/presentation/api/admin; then
                    echo -e "${GREEN}✓ Integration tests passed${NC}"

                    # Generate integration coverage
                    echo -e "${YELLOW}📊 Generating integration test coverage...${NC}"
                    go test -race -tags=integration -p 1 -count=1 \
                        -covermode=atomic -coverprofile="$COVERAGE_DIR/backend-integration.out" \
                        ./backend/internal/infrastructure/database/... ./backend/internal/presentation/api/admin 2>&1 | grep -v "no test files" || true
                    echo -e "${GREEN}✓ Integration coverage generated${NC}"
                else
                    echo -e "${RED}❌ Integration tests failed${NC}"
                    FAILED=$((FAILED + 1))
                fi
            else
                echo -e "${RED}❌ Migrations failed${NC}"
                FAILED=$((FAILED + 1))
                INTEGRATION_SKIPPED=1
            fi
        fi
    else
        echo -e "${RED}❌ Failed to start Docker services${NC}"
        FAILED=$((FAILED + 1))
        INTEGRATION_SKIPPED=1
    fi
fi
echo ""

# Merge backend coverage
echo -e "${YELLOW}📊 Merging backend coverage reports...${NC}"
echo "mode: atomic" > "$COVERAGE_DIR/backend-coverage.out"
tail -n +2 "$COVERAGE_DIR/backend-unit.out" >> "$COVERAGE_DIR/backend-coverage.out" 2>/dev/null || true
if [ "$INTEGRATION_SKIPPED" = "0" ] && [ -f "$COVERAGE_DIR/backend-integration.out" ]; then
    tail -n +2 "$COVERAGE_DIR/backend-integration.out" >> "$COVERAGE_DIR/backend-coverage.out" 2>/dev/null || true
fi

# Extract backend coverage percentage
BACKEND_COV=$(go tool cover -func="$COVERAGE_DIR/backend-coverage.out" 2>/dev/null | tail -1 | awk '{print $3}' || echo "N/A")
echo -e "${GREEN}✓ Backend coverage: $BACKEND_COV${NC}"
echo ""

# Cleanup integration environment
cleanup_integration
INTEGRATION_STARTED=0

# ==============================================================================
# Phase 2: Frontend Unit Tests
# ==============================================================================
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Phase 2/3: Frontend Unit Tests (Vitest)${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$WEBAPP_DIR"

echo -e "${YELLOW}📦 Installing frontend dependencies...${NC}"
if npm ci --no-audit --no-fund --prefer-offline > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Dependencies installed${NC}"
else
    echo -e "${RED}❌ Failed to install dependencies${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

echo -e "${YELLOW}🧪 Running frontend unit tests with coverage...${NC}"
if npm run test:coverage; then
    echo -e "${GREEN}✓ Frontend unit tests passed${NC}"

    # Extract frontend coverage percentage from lcov.info
    if [ -f "coverage/lcov.info" ]; then
        cp coverage/lcov.info "$COVERAGE_DIR/frontend-unit.lcov"

        # Calculate coverage from lcov.info
        FRONTEND_LINES_FOUND=$(grep -c "^DA:" coverage/lcov.info 2>/dev/null || echo "0")
        FRONTEND_LINES_MISS=$(grep "^DA:" coverage/lcov.info 2>/dev/null | grep -c ",0$" || echo "0")
        # Remove any whitespace
        FRONTEND_LINES_FOUND=$(echo "$FRONTEND_LINES_FOUND" | tr -d '[:space:]')
        FRONTEND_LINES_MISS=$(echo "$FRONTEND_LINES_MISS" | tr -d '[:space:]')
        # Default to 0 if empty
        FRONTEND_LINES_FOUND=${FRONTEND_LINES_FOUND:-0}
        FRONTEND_LINES_MISS=${FRONTEND_LINES_MISS:-0}
        FRONTEND_LINES_HIT=$((FRONTEND_LINES_FOUND - FRONTEND_LINES_MISS))
        if [ "$FRONTEND_LINES_FOUND" -gt 0 ]; then
            FRONTEND_COV=$(awk "BEGIN {printf \"%.1f%%\", ($FRONTEND_LINES_HIT/$FRONTEND_LINES_FOUND)*100}")
        else
            FRONTEND_COV="0.0%"
        fi
        echo -e "${GREEN}✓ Frontend unit coverage: $FRONTEND_COV${NC}"
    else
        echo -e "${YELLOW}⚠️  Coverage file not found${NC}"
        FRONTEND_COV="N/A"
    fi
else
    echo -e "${RED}❌ Frontend unit tests failed${NC}"
    FAILED=$((FAILED + 1))
    FRONTEND_COV="N/A"
fi
echo ""

# ==============================================================================
# Phase 3: E2E Tests (Cypress)
# ==============================================================================
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Phase 3/3: E2E Tests (Cypress)${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$PROJECT_ROOT"

# Update trap to cleanup E2E instead
trap cleanup_e2e EXIT

if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️  Docker not available, skipping E2E tests${NC}"
    E2E_SKIPPED=1
    E2E_COV="N/A"
else
    echo -e "${YELLOW}🧹 Cleaning up previous E2E environment...${NC}"
    docker compose -f compose.e2e.yml down -v --remove-orphans > /dev/null 2>&1 || true
    echo -e "${GREEN}✓ Cleanup complete${NC}"
    echo ""

    echo -e "${YELLOW}🏗️  Building frontend with coverage instrumentation...${NC}"
    cd "$WEBAPP_DIR"
    if CYPRESS_COVERAGE=true npm run build > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Frontend built with instrumentation${NC}"
    else
        echo -e "${RED}❌ Failed to build frontend${NC}"
        FAILED=$((FAILED + 1))
        E2E_SKIPPED=1
    fi
    echo ""

    if [ "$E2E_SKIPPED" = "0" ]; then
        cd "$PROJECT_ROOT"
        echo -e "${YELLOW}🐳 Starting E2E stack (compose.e2e.yml --build)...${NC}"
        if docker compose -f compose.e2e.yml up -d --force-recreate --build; then
            E2E_STARTED=1
            echo -e "${GREEN}✓ E2E stack started${NC}"
            echo ""

            # Wait for services
            echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
            echo "   - Waiting for database..."
            sleep 5

            echo "   - Waiting for migrations..."
            docker compose -f compose.e2e.yml logs ackify-migrate 2>&1 | tail -5

            echo "   - Waiting for backend..."
            sleep 10

            # Health check
            MAX_RETRIES=30
            RETRY_COUNT=0
            while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
                if curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
                    echo -e "${GREEN}✓ Backend is ready!${NC}"
                    break
                fi
                RETRY_COUNT=$((RETRY_COUNT + 1))
                echo "   Retry $RETRY_COUNT/$MAX_RETRIES..."
                sleep 2
            done

            if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
                echo -e "${RED}❌ Backend failed to start. Check logs:${NC}"
                docker compose -f compose.e2e.yml logs ackify-ce | tail -50
                FAILED=$((FAILED + 1))
                E2E_SKIPPED=1
            else
                echo -e "${GREEN}✓ All services are ready${NC}"
                echo ""

                # Run Cypress tests
                cd "$WEBAPP_DIR"
                echo -e "${YELLOW}🧪 Running Cypress E2E tests...${NC}"
                if npm run test:e2e; then
                    echo -e "${GREEN}✓ E2E tests passed${NC}"

                    # Extract E2E coverage
                    if [ -f "coverage-e2e/lcov.info" ]; then
                        cp coverage-e2e/lcov.info "$COVERAGE_DIR/e2e.lcov"

                        # Calculate coverage from lcov.info
                        E2E_LINES_FOUND=$(grep -c "^DA:" coverage-e2e/lcov.info 2>/dev/null || echo "0")
                        E2E_LINES_MISS=$(grep "^DA:" coverage-e2e/lcov.info 2>/dev/null | grep -c ",0$" || echo "0")
                        # Remove any whitespace
                        E2E_LINES_FOUND=$(echo "$E2E_LINES_FOUND" | tr -d '[:space:]')
                        E2E_LINES_MISS=$(echo "$E2E_LINES_MISS" | tr -d '[:space:]')
                        # Default to 0 if empty
                        E2E_LINES_FOUND=${E2E_LINES_FOUND:-0}
                        E2E_LINES_MISS=${E2E_LINES_MISS:-0}
                        E2E_LINES_HIT=$((E2E_LINES_FOUND - E2E_LINES_MISS))
                        if [ "$E2E_LINES_FOUND" -gt 0 ]; then
                            E2E_COV=$(awk "BEGIN {printf \"%.1f%%\", ($E2E_LINES_HIT/$E2E_LINES_FOUND)*100}")
                        else
                            E2E_COV="0.0%"
                        fi
                        echo -e "${GREEN}✓ E2E coverage: $E2E_COV${NC}"
                    else
                        echo -e "${YELLOW}⚠️  E2E coverage file not found${NC}"
                        E2E_COV="N/A"
                    fi
                else
                    echo -e "${RED}❌ E2E tests failed${NC}"
                    FAILED=$((FAILED + 1))
                    E2E_COV="N/A"
                fi
            fi
        else
            echo -e "${RED}❌ Failed to start E2E stack${NC}"
            FAILED=$((FAILED + 1))
            E2E_SKIPPED=1
            E2E_COV="N/A"
        fi
    fi
fi
echo ""

# Cleanup E2E environment
cleanup_e2e
E2E_STARTED=0

# ==============================================================================
# Summary
# ==============================================================================
cd "$PROJECT_ROOT"

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                   Coverage Summary                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${CYAN}Backend (Go):${NC}          $BACKEND_COV"
echo -e "  ${CYAN}Frontend (Vitest):${NC}     $FRONTEND_COV"
echo -e "  ${CYAN}E2E (Cypress):${NC}         $E2E_COV"
echo ""

if [ "$INTEGRATION_SKIPPED" = "1" ]; then
    echo -e "${YELLOW}⚠️  Integration tests were skipped${NC}"
fi
if [ "$E2E_SKIPPED" = "1" ]; then
    echo -e "${YELLOW}⚠️  E2E tests were skipped${NC}"
fi
echo ""

echo -e "${BLUE}Coverage reports saved to:${NC} $COVERAGE_DIR"
echo ""

# Display coverage files
if [ -d "$COVERAGE_DIR" ]; then
    echo -e "${BLUE}Generated files:${NC}"
    ls -lh "$COVERAGE_DIR" 2>/dev/null | tail -n +2 | awk '{print "  - " $9 " (" $5 ")"}'
    echo ""
fi

# Final result
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}║  ✓ All test suites passed successfully!                  ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}║  ✗ $FAILED test suite(s) failed                              ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
