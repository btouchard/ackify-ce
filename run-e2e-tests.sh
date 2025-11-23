#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Script to run Ackify E2E tests

set -e

echo "🚀 Ackify E2E Test Runner"
echo "=========================="
echo ""

TMP_PATH=${PWD}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if docker compose is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Step 1: Clean up any existing containers/volumes
echo -e "${YELLOW}🧹 Cleaning up previous test environment...${NC}"
docker compose -f compose.e2e.yml down -v 2>/dev/null || true
echo -e "${GREEN}✓ Cleanup complete${NC}"
echo ""

# Step 2: Start the stack
echo -e "${YELLOW}🐳 Starting E2E stack (backend + PostgreSQL + Mailhog)...${NC}"
docker compose -f compose.e2e.yml down -v
docker compose -f compose.e2e.yml up -d --force-recreate --build

# Step 3: Wait for services to be ready
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
    docker compose -f compose.e2e.yml logs ackify-ce
    exit 1
fi

echo -e "${GREEN}✓ All services are ready${NC}"
echo ""

# Step 4: Run tests
echo -e "${YELLOW}🧪 Running Cypress E2E tests...${NC}"
echo ""

cd webapp

if [ "$1" == "open" ]; then
    echo -e "${BLUE}Opening Cypress interactive mode...${NC}"
    npm run test:e2e:open
else
    echo -e "${BLUE}Running tests in headless mode...${NC}"
    npm run test:e2e
    TEST_EXIT_CODE=$?

    cd ..

    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✅ All tests passed!${NC}"
    else
        echo ""
        echo -e "${RED}❌ Some tests failed. Check the output above.${NC}"
    fi
fi

cd "${TMP_PATH}"

# Step 5: Cleanup prompt
echo ""
echo -e "${YELLOW}🧹 Cleanup${NC}"
read -p "Do you want to clean up the test environment? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Stopping and removing containers...${NC}"
    docker compose -f compose.e2e.yml down -v
    echo -e "${GREEN}✓ Cleanup complete${NC}"
else
    echo -e "${BLUE}ℹ️  Test environment is still running.${NC}"
    echo "   - Backend: http://localhost:8080"
    echo "   - Mailhog: http://localhost:8025"
    echo ""
    echo "   To clean up later, run:"
    echo "   docker compose -f compose.e2e.yml down -v"
fi

echo ""
echo -e "${GREEN}Done!${NC}"

exit ${TEST_EXIT_CODE:-0}
