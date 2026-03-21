#!/bin/bash

# E2E Test Script for Claude Pipeline
# Tests complete workflows from API to execution

set -e

BASE_URL="${1:-http://localhost:8080}"
TIMEOUT="${2:-60}"

echo "========================================"
echo "Claude Pipeline E2E Tests"
echo "========================================"
echo "Base URL: $BASE_URL"
echo "Timeout: ${TIMEOUT}s"
echo ""

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
test_case() {
    local name="$1"
    echo ""
    echo "TEST: $name"
    echo "---"
    ((TESTS_RUN++))
}

pass() {
    echo "✓ PASSED"
    ((TESTS_PASSED++))
}

fail() {
    local reason="$1"
    echo "✗ FAILED: $reason"
    ((TESTS_FAILED++))
}

check_response() {
    local expected="$1"
    local actual="$2"

    if echo "$actual" | grep -q "$expected"; then
        return 0
    else
        return 1
    fi
}

wait_for_task() {
    local task_id="$1"
    local max_wait="${2:-30}"
    local elapsed=0

    while [ $elapsed -lt $max_wait ]; do
        local status=$(curl -sf "$BASE_URL/api/tasks/$task_id" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

        if [ "$status" = "completed" ] || [ "$status" = "failed" ]; then
            echo "$status"
            return 0
        fi

        sleep 1
        ((elapsed++))
    done

    echo "timeout"
    return 1
}

# Check server availability
test_case "Server Health Check"
if curl -sf "$BASE_URL/health" > /dev/null; then
    pass
else
    fail "Server not responding"
    echo "Make sure the server is running: ./bin/server"
    exit 1
fi

# Test 1: List Skills
test_case "List Available Skills"
RESPONSE=$(curl -sf "$BASE_URL/api/skills")
if check_response "skills" "$RESPONSE"; then
    SKILL_COUNT=$(echo "$RESPONSE" | grep -o '"id"' | wc -l)
    echo "Found $SKILL_COUNT skills"
    pass
else
    fail "No skills returned"
fi

# Test 2: Get Specific Skill
test_case "Get Skill Details"
RESPONSE=$(curl -sf "$BASE_URL/api/skills/code-review")
if check_response "code-review" "$RESPONSE"; then
    echo "Skill: $(echo "$RESPONSE" | grep -o '"name":"[^"]*"' | head -1)"
    pass
else
    fail "Skill not found"
fi

# Test 3: Create Task
test_case "Create Task"
RESPONSE=$(curl -sf -X POST "$BASE_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{
        "skill_id": "code-review",
        "parameters": {
            "target": "src/",
            "depth": "quick"
        }
    }')

if check_response "task-" "$RESPONSE"; then
    TASK_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "Created task: $TASK_ID"
    pass
else
    fail "Task creation failed: $RESPONSE"
fi

# Test 4: Get Task Status
test_case "Get Task Status"
if [ -n "$TASK_ID" ]; then
    RESPONSE=$(curl -sf "$BASE_URL/api/tasks/$TASK_ID")
    if check_response "$TASK_ID" "$RESPONSE"; then
        STATUS=$(echo "$RESPONSE" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        echo "Task status: $STATUS"
        pass
    else
        fail "Could not get task status"
    fi
else
    fail "No task ID available"
fi

# Test 5: List Tasks
test_case "List All Tasks"
RESPONSE=$(curl -sf "$BASE_URL/api/tasks")
if check_response "tasks" "$RESPONSE"; then
    TASK_COUNT=$(echo "$RESPONSE" | grep -o '"id"' | wc -l)
    echo "Found $TASK_COUNT tasks"
    pass
else
    fail "No tasks returned"
fi

# Test 6: Create Pipeline
test_case "Create Pipeline"
RESPONSE=$(curl -sf -X POST "$BASE_URL/api/pipelines" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "e2e-test-pipeline",
        "description": "E2E test pipeline",
        "mode": "serial",
        "steps": [
            {
                "id": "step-1",
                "name": "Echo Test",
                "cli": "echo",
                "action": "command",
                "command": "Hello E2E Test",
                "params": {}
            }
        ]
    }')

if check_response "pipeline-" "$RESPONSE"; then
    PIPELINE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "Created pipeline: $PIPELINE_ID"
    pass
else
    fail "Pipeline creation failed: $RESPONSE"
fi

# Test 7: Get Pipeline
test_case "Get Pipeline Details"
if [ -n "$PIPELINE_ID" ]; then
    RESPONSE=$(curl -sf "$BASE_URL/api/pipelines/$PIPELINE_ID")
    if check_response "$PIPELINE_ID" "$RESPONSE"; then
        echo "Pipeline name: $(echo "$RESPONSE" | grep -o '"name":"[^"]*"' | head -1)"
        pass
    else
        fail "Could not get pipeline"
    fi
else
    fail "No pipeline ID available"
fi

# Test 8: List Pipelines
test_case "List All Pipelines"
RESPONSE=$(curl -sf "$BASE_URL/api/pipelines")
if check_response "pipelines" "$RESPONSE"; then
    PIPELINE_COUNT=$(echo "$RESPONSE" | grep -o '"id"' | wc -l)
    echo "Found $PIPELINE_COUNT pipelines"
    pass
else
    fail "No pipelines returned"
fi

# Test 9: Run Pipeline
test_case "Run Pipeline"
if [ -n "$PIPELINE_ID" ]; then
    RESPONSE=$(curl -sf -X POST "$BASE_URL/api/pipelines/$PIPELINE_ID/run" \
        -H "Content-Type: application/json" \
        -d '{}')

    if check_response "run-" "$RESPONSE"; then
        RUN_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        echo "Started run: $RUN_ID"
        pass
    else
        fail "Pipeline run failed: $RESPONSE"
    fi
else
    fail "No pipeline ID available"
fi

# Test 10: Get Run Status
test_case "Get Run Status"
if [ -n "$RUN_ID" ]; then
    RESPONSE=$(curl -sf "$BASE_URL/api/runs/$RUN_ID")
    if check_response "$RUN_ID" "$RESPONSE"; then
        STATUS=$(echo "$RESPONSE" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        echo "Run status: $STATUS"
        pass
    else
        fail "Could not get run status"
    fi
else
    fail "No run ID available"
fi

# Test 11: Service Status
test_case "Service Status"
RESPONSE=$(curl -sf "$BASE_URL/api/status")
if check_response "healthy" "$RESPONSE"; then
    echo "Service is healthy"
    pass
else
    fail "Service unhealthy"
fi

# Test 12: Delete Pipeline
test_case "Delete Pipeline"
if [ -n "$PIPELINE_ID" ]; then
    RESPONSE=$(curl -sf -X DELETE "$BASE_URL/api/pipelines/$PIPELINE_ID")
    if check_response "deleted" "$RESPONSE" || [ -z "$RESPONSE" ]; then
        echo "Pipeline deleted"
        pass
    else
        fail "Delete failed: $RESPONSE"
    fi
else
    fail "No pipeline ID available"
fi

# Test 13: Error Handling - Invalid Skill
test_case "Error Handling: Invalid Skill ID"
RESPONSE=$(curl -sf -X POST "$BASE_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{"skill_id": "non-existent-skill", "parameters": {}}' 2>&1 || true)

if echo "$RESPONSE" | grep -q "error\|not found"; then
    echo "Error correctly returned for invalid skill"
    pass
else
    fail "Should have returned error for invalid skill"
fi

# Test 14: Error Handling - Missing Required Field
test_case "Error Handling: Missing Required Field"
RESPONSE=$(curl -sf -X POST "$BASE_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{"parameters": {}}' 2>&1 || true)

if echo "$RESPONSE" | grep -q "error\|required"; then
    echo "Error correctly returned for missing skill_id"
    pass
else
    fail "Should have returned error for missing required field"
fi

# Summary
echo ""
echo "========================================"
echo "E2E Test Summary"
echo "========================================"
echo "Tests Run:    $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -gt 0 ]; then
    echo "Some tests failed!"
    exit 1
else
    echo "All E2E tests passed! ✓"
    exit 0
fi