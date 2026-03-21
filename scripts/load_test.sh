#!/bin/bash

# Load Test Script for Claude Pipeline API
# Simulates concurrent users and measures performance

set -e

BASE_URL="${1:-http://localhost:8080}"
NUM_REQUESTS="${2:-100}"
CONCURRENT="${3:-10}"

echo "Claude Pipeline Load Test"
echo "========================"
echo "Base URL: $BASE_URL"
echo "Total Requests: $NUM_REQUESTS"
echo "Concurrent Users: $CONCURRENT"
echo ""

# Check if server is running
if ! curl -sf "$BASE_URL/health" > /dev/null; then
    echo "Error: Server is not running at $BASE_URL"
    exit 1
fi

# Function to make requests and measure time
make_request() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    local start=$(date +%s%N)

    if [ "$method" = "GET" ]; then
        curl -sf "$BASE_URL$endpoint" > /dev/null
    else
        curl -sf -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" > /dev/null
    fi

    local end=$(date +%s%N)
    local duration=$(( (end - start) / 1000000 ))
    echo $duration
}

# Test endpoints
echo "Testing individual endpoints..."
echo ""

# Warmup
echo "Warmup..."
for i in {1..10}; do
    curl -sf "$BASE_URL/api/status" > /dev/null
done

# Test GET /api/skills
echo "GET /api/skills:"
TOTAL=0
for i in $(seq 1 $NUM_REQUESTS); do
    TIME=$(make_request "/api/skills" "GET")
    TOTAL=$((TOTAL + TIME))
done
AVG=$((TOTAL / NUM_REQUESTS))
echo "  Average: ${AVG}ms"

# Test GET /api/status
echo "GET /api/status:"
TOTAL=0
for i in $(seq 1 $NUM_REQUESTS); do
    TIME=$(make_request "/api/status" "GET")
    TOTAL=$((TOTAL + TIME))
done
AVG=$((TOTAL / NUM_REQUESTS))
echo "  Average: ${AVG}ms"

# Test POST /api/tasks
echo "POST /api/tasks:"
TOTAL=0
for i in $(seq 1 $NUM_REQUESTS); do
    TIME=$(make_request "/api/tasks" "POST" '{"skill_id":"code-review","parameters":{"target":"src/","depth":"quick"}}')
    TOTAL=$((TOTAL + TIME))
done
AVG=$((TOTAL / NUM_REQUESTS))
echo "  Average: ${AVG}ms"

# Test POST /api/pipelines
echo "POST /api/pipelines:"
TOTAL=0
for i in $(seq 1 $NUM_REQUESTS); do
    TIME=$(make_request "/api/pipelines" "POST" '{"name":"load-test","description":"Load test","mode":"serial","steps":[{"id":"s1","name":"Step","cli":"echo","action":"cmd","params":{}}]}')
    TOTAL=$((TOTAL + TIME))
done
AVG=$((TOTAL / NUM_REQUESTS))
echo "  Average: ${AVG}ms"

# Concurrent test
echo ""
echo "Concurrent load test ($CONCURRENT concurrent requests)..."

PIDS=""
START=$(date +%s%N)

for i in $(seq 1 $CONCURRENT); do
    (
        for j in $(seq 1 $((NUM_REQUESTS / CONCURRENT))); do
            curl -sf "$BASE_URL/api/status" > /dev/null
        done
    ) &
    PIDS="$PIDS $!"
done

# Wait for all background processes
wait $PIDS

END=$(date +%s%N)
DURATION=$(( (END - START) / 1000000 ))
RPS=$(( (NUM_REQUESTS * 1000) / DURATION ))

echo "  Total time: ${DURATION}ms"
echo "  Requests/second: ${RPS}"

echo ""
echo "Load test completed!"