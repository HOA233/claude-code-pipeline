#!/bin/bash

# Comprehensive Test Suite for Claude Pipeline
# Runs unit tests, integration tests, benchmarks, and API tests

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Claude Pipeline - Test Suite           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Check dependencies
echo -e "${YELLOW}Checking dependencies...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go $(go version | awk '{print $3}')${NC}"

# Check Redis
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠ Redis is not running. Some tests will be skipped.${NC}"
    REDIS_AVAILABLE=false
else
    echo -e "${GREEN}✓ Redis is running${NC}"
    REDIS_AVAILABLE=true
fi

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run tests and track results
run_test() {
    local name="$1"
    local cmd="$2"

    echo ""
    echo -e "${YELLOW}Running: $name${NC}"

    if eval "$cmd"; then
        echo -e "${GREEN}✓ $name passed${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ $name failed${NC}"
        ((TESTS_FAILED++))
    fi
}

# 1. Code Quality
echo ""
echo -e "${BLUE}═══ Code Quality ═══${NC}"

run_test "Go Vet" "go vet ./..."
run_test "Go Fmt" "test -z \"\$(gofmt -l .)\""
run_test "Go Mod Tidy" "go mod tidy && git diff --exit-code go.mod go.sum"

# 2. Unit Tests
echo ""
echo -e "${BLUE}═══ Unit Tests ═══${NC}"

run_test "Unit Tests" "go test -v -short ./..."

# 3. Integration Tests (requires Redis)
if [ "$REDIS_AVAILABLE" = true ]; then
    echo ""
    echo -e "${BLUE}═══ Integration Tests ═══${NC}"

    run_test "Integration Tests" "go test -v -run Integration ./tests/..."
fi

# 4. Build
echo ""
echo -e "${BLUE}═══ Build ═══${NC}"

run_test "Build Binary" "go build -o bin/server ./cmd/server"

# 5. API Tests (requires running server)
if [ "$REDIS_AVAILABLE" = true ]; then
    echo ""
    echo -e "${BLUE}═══ API Tests ═══${NC}"

    # Start server in background
    echo "Starting test server..."
    ./bin/server &
    SERVER_PID=$!
    sleep 3

    # Cleanup function
    cleanup() {
        echo ""
        echo "Stopping test server..."
        kill $SERVER_PID 2>/dev/null || true
    }
    trap cleanup EXIT

    BASE_URL="http://localhost:8080"

    # Health check
    if curl -sf "$BASE_URL/health" > /dev/null; then
        echo -e "${GREEN}✓ Health endpoint working${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ Health endpoint failed${NC}"
        ((TESTS_FAILED++))
    fi

    # Skills endpoint
    if curl -sf "$BASE_URL/api/skills" > /dev/null; then
        echo -e "${GREEN}✓ Skills endpoint working${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ Skills endpoint failed${NC}"
        ((TESTS_FAILED++))
    fi

    # Create task
    TASK_RESPONSE=$(curl -sf -X POST "$BASE_URL/api/tasks" \
        -H "Content-Type: application/json" \
        -d '{"skill_id":"code-review","parameters":{"target":"src/","depth":"quick"}}')

    if echo "$TASK_RESPONSE" | grep -q "task-"; then
        echo -e "${GREEN}✓ Task creation working${NC}"
        ((TESTS_PASSED++))
        TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

        # Get task
        if curl -sf "$BASE_URL/api/tasks/$TASK_ID" > /dev/null; then
            echo -e "${GREEN}✓ Task retrieval working${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}✗ Task retrieval failed${NC}"
            ((TESTS_FAILED++))
        fi
    else
        echo -e "${RED}✗ Task creation failed${NC}"
        ((TESTS_FAILED++))
    fi

    # Create pipeline
    PIPELINE_RESPONSE=$(curl -sf -X POST "$BASE_URL/api/pipelines" \
        -H "Content-Type: application/json" \
        -d '{"name":"test-pipeline","description":"Test","mode":"serial","steps":[{"id":"s1","name":"Step1","cli":"echo","action":"command","command":"hello","params":{}}]}')

    if echo "$PIPELINE_RESPONSE" | grep -q "pipeline-"; then
        echo -e "${GREEN}✓ Pipeline creation working${NC}"
        ((TESTS_PASSED++))
        PIPELINE_ID=$(echo "$PIPELINE_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

        # Delete pipeline
        if curl -sf -X DELETE "$BASE_URL/api/pipelines/$PIPELINE_ID" > /dev/null; then
            echo -e "${GREEN}✓ Pipeline deletion working${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}✗ Pipeline deletion failed${NC}"
            ((TESTS_FAILED++))
        fi
    else
        echo -e "${RED}✗ Pipeline creation failed${NC}"
        ((TESTS_FAILED++))
    fi
fi

# 6. Benchmarks (optional)
if [ "$1" = "--bench" ]; then
    echo ""
    echo -e "${BLUE}═══ Benchmarks ═══${NC}"

    run_test "Benchmarks" "go test -bench=. -benchmem ./tests/..."
fi

# Summary
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              Test Results                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
fi