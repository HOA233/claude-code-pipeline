#!/bin/bash
# Load Testing Script for Claude Pipeline API
# Uses curl for HTTP load testing

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
CONCURRENT="${CONCURRENT:-10}"
REQUESTS="${REQUESTS:-100}"

echo "============================================"
echo "Claude Pipeline Load Test"
echo "============================================"
echo "Base URL: $BASE_URL"
echo "Concurrent: $CONCURRENT"
echo "Total Requests: $REQUESTS"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Statistics
TOTAL=0
SUCCESS=0
FAILED=0
TOTAL_TIME=0

# Test endpoint with timing
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3

    start=$(date +%s%N)

    if [ -z "$data" ]; then
        response=$(curl -sf -w "\n%{http_code}" -X "$method" "${BASE_URL}${endpoint}" 2>/dev/null)
    else
        response=$(curl -sf -w "\n%{http_code}" -X "$method" "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null)
    fi

    end=$(date +%s%N)
    duration=$(( (end - start) / 1000000 ))

    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✓${NC} $method $endpoint (${duration}ms)"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} $method $endpoint - HTTP $http_code"
        ((FAILED++))
    fi

    ((TOTAL++))
    TOTAL_TIME=$((TOTAL_TIME + duration))
}

echo "=== Phase 1: Baseline Tests ==="
echo ""

# Health check
test_endpoint "GET" "/health"
test_endpoint "GET" "/ready"
test_endpoint "GET" "/live"

echo ""
echo "=== Phase 2: Skills Endpoints ==="
echo ""

for i in $(seq 1 5); do
    test_endpoint "GET" "/api/skills"
done

test_endpoint "GET" "/api/skills/code-review"
test_endpoint "GET" "/api/skills/deploy"

echo ""
echo "=== Phase 3: Tasks Endpoints ==="
echo ""

# Create tasks
for i in $(seq 1 5); do
    test_endpoint "POST" "/api/tasks" '{
        "skill_id": "code-review",
        "parameters": {
            "target": "src/",
            "depth": "quick"
        }
    }'
done

# List tasks
for i in $(seq 1 5); do
    test_endpoint "GET" "/api/tasks"
done

echo ""
echo "=== Phase 4: Pipelines Endpoints ==="
echo ""

# Create pipelines
for i in $(seq 1 3); do
    test_endpoint "POST" "/api/pipelines" '{
        "name": "load-test-pipeline",
        "mode": "serial",
        "steps": [
            {"id": "step1", "cli": "echo", "action": "test", "params": {}}
        ]
    }'
done

# List pipelines
for i in $(seq 1 5); do
    test_endpoint "GET" "/api/pipelines"
done

echo ""
echo "=== Phase 5: Concurrent Load Test ==="
echo ""

# Function for concurrent testing
concurrent_test() {
    local endpoint=$1
    local count=$2

    echo "Testing $endpoint with $count concurrent requests..."

    for i in $(seq 1 "$count"); do
        (
            start=$(date +%s%N)
            curl -sf "${BASE_URL}${endpoint}" -o /dev/null -w "%{http_code}" 2>/dev/null
            end=$(date +%s%N)
            echo "$(( (end - start) / 1000000 ))"
        ) &
    done

    wait
}

# Run concurrent tests
concurrent_test "/api/skills" "$CONCURRENT"
concurrent_test "/api/tasks" "$CONCURRENT"
concurrent_test "/api/pipelines" "$CONCURRENT"
concurrent_test "/api/status" "$CONCURRENT"

echo ""
echo "=== Phase 6: Stress Test ==="
echo ""

echo "Running $REQUESTS requests to /health..."

for i in $(seq 1 "$REQUESTS"); do
    curl -sf "${BASE_URL}/health" -o /dev/null -w "." 2>/dev/null
done

echo ""
echo ""

echo "============================================"
echo "Load Test Summary"
echo "============================================"
echo ""
echo "Total Requests: $TOTAL"
echo -e "Successful:     ${GREEN}$SUCCESS${NC}"
echo -e "Failed:         ${RED}$FAILED${NC}"

if [ $TOTAL -gt 0 ]; then
    AVG_TIME=$((TOTAL_TIME / TOTAL))
    echo "Avg Latency:    ${AVG_TIME}ms"
fi

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}Some tests failed. Check logs for details.${NC}"
    exit 1
fi