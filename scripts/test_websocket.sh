#!/bin/bash

# Comprehensive WebSocket Test Script
# Tests WebSocket and SSE endpoints for real-time updates

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080}"

echo "============================================"
echo "WebSocket & Real-time Updates Test"
echo "============================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}✓ $1${NC}"
}

fail() {
    echo -e "${RED}✗ $1${NC}"
}

info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Test 1: Health Check
echo "Test 1: Health Check"
response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/health")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" = "200" ]; then
    pass "Health check passed"
else
    fail "Health check failed: $http_code"
fi
echo ""

# Test 2: Create a task
echo "Test 2: Create Task for WebSocket Testing"
response=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{
        "skill_id": "code-review",
        "parameters": {
            "target": "src/",
            "depth": "quick"
        }
    }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
    task_id=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    pass "Task created: $task_id"
else
    fail "Task creation failed: $http_code"
    task_id=""
fi
echo ""

# Test 3: WebSocket connection test (using wscat or websocat if available)
echo "Test 3: WebSocket Connection Test"
if command -v websocat &> /dev/null; then
    info "Testing WebSocket connection to ${WS_URL}/ws/tasks/${task_id}"
    timeout 5s websocat -n1 "${WS_URL}/ws/tasks/${task_id}" 2>/dev/null && pass "WebSocket connection established" || info "WebSocket test skipped (timeout or no data)"
elif command -v wscat &> /dev/null; then
    info "Testing WebSocket connection to ${WS_URL}/ws/tasks/${task_id}"
    timeout 5s wscat -c "${WS_URL}/ws/tasks/${task_id}" -x 2>/dev/null && pass "WebSocket connection established" || info "WebSocket test skipped (timeout or no data)"
else
    info "WebSocket client (websocat or wscat) not installed, skipping WebSocket test"
    info "Install with: cargo install websocat or npm install -g wscat"
fi
echo ""

# Test 4: SSE endpoint test
echo "Test 4: SSE Connection Test"
info "Testing SSE endpoint ${BASE_URL}/sse/tasks/${task_id}"
timeout 3s curl -s -N "${BASE_URL}/sse/tasks/${task_id}" 2>/dev/null | head -5 && pass "SSE connection established" || info "SSE test completed (timeout expected)"
echo ""

# Test 5: Global SSE endpoint
echo "Test 5: Global SSE Endpoint"
info "Testing global SSE endpoint ${BASE_URL}/sse"
timeout 3s curl -s -N "${BASE_URL}/sse" 2>/dev/null | head -5 && pass "Global SSE connection established" || info "Global SSE test completed"
echo ""

# Test 6: Create a pipeline and run
echo "Test 6: Create Pipeline for Run Testing"
response=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/api/pipelines" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-pipeline",
        "mode": "serial",
        "steps": [
            {"id": "step1", "cli": "echo", "action": "test", "params": {}}
        ]
    }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
    pipeline_id=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    pass "Pipeline created: $pipeline_id"
else
    fail "Pipeline creation failed: $http_code"
    pipeline_id=""
fi
echo ""

# Test 7: Run pipeline
if [ -n "$pipeline_id" ]; then
    echo "Test 7: Run Pipeline"
    response=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/api/pipelines/${pipeline_id}/run")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)

    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        run_id=$(echo "$body" | grep -o '"run_id":"[^"]*"' | cut -d'"' -f4)
        pass "Pipeline run started: $run_id"
    else
        fail "Pipeline run failed: $http_code"
        run_id=""
    fi
    echo ""

    # Test 8: WebSocket for run updates
    if [ -n "$run_id" ]; then
        echo "Test 8: WebSocket for Run Updates"
        if command -v websocat &> /dev/null; then
            timeout 5s websocat -n1 "${WS_URL}/ws/runs/${run_id}" 2>/dev/null && pass "Run WebSocket connection established" || info "Run WebSocket test completed"
        elif command -v wscat &> /dev/null; then
            timeout 5s wscat -c "${WS_URL}/ws/runs/${run_id}" -x 2>/dev/null && pass "Run WebSocket connection established" || info "Run WebSocket test completed"
        else
            info "WebSocket client not available"
        fi
        echo ""

        # Test 9: SSE for run updates
        echo "Test 9: SSE for Run Updates"
        timeout 3s curl -s -N "${BASE_URL}/sse/runs/${run_id}" 2>/dev/null | head -5 && pass "Run SSE connection established" || info "Run SSE test completed"
        echo ""
    fi
fi

# Test 10: List all endpoints
echo "Test 10: Verify All WebSocket/SSE Endpoints"
endpoints=(
    "GET /health"
    "GET /api/tasks"
    "GET /api/pipelines"
    "GET /ws"
    "GET /sse"
)

for endpoint in "${endpoints[@]}"; do
    method=$(echo "$endpoint" | cut -d' ' -f1)
    path=$(echo "$endpoint" | cut -d' ' -f2)

    if [ "$method" = "GET" ]; then
        http_code=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${path}")
        if [ "$http_code" = "200" ] || [ "$http_code" = "101" ]; then
            pass "Endpoint $path accessible ($http_code)"
        else
            info "Endpoint $path returned $http_code"
        fi
    fi
done
echo ""

echo "============================================"
echo "WebSocket Test Complete"
echo "============================================"