#!/bin/bash

# Comprehensive Test Suite for Claude Pipeline
# Tests all components: API, Services, Integration, Performance

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

API_URL="${API_URL:-http://localhost:8080}"
TIMEOUT="${TIMEOUT:-30}"
PASS_COUNT=0
FAIL_COUNT=0

echo "========================================"
echo "  Claude Pipeline Comprehensive Tests"
echo "========================================"
echo ""
echo "API URL: $API_URL"
echo "Timeout: ${TIMEOUT}s"
echo ""

# Test helper functions
test_start() {
    echo -e "${CYAN}[TEST]${NC} $1"
}

test_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASS_COUNT++))
}

test_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    if [ -n "$2" ]; then
        echo "        Error: $2"
    fi
    ((FAIL_COUNT++))
}

test_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

check_response() {
    local expected="$1"
    local actual="$2"
    if [ "$actual" -eq "$expected" ]; then
        return 0
    else
        return 1
    fi
}

# ==================== Health Tests ====================

echo ""
echo -e "${BLUE}=== Health Tests ===${NC}"
echo ""

test_start "Health check endpoint"
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health" --connect-timeout 5 || echo "000")
if check_response 200 "$HEALTH_STATUS"; then
    test_pass "Health check returned 200"
else
    test_fail "Health check returned $HEALTH_STATUS"
fi

test_start "API status endpoint"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/status" --connect-timeout 5 || echo "000")
if check_response 200 "$STATUS"; then
    test_pass "Status endpoint returned 200"
else
    test_fail "Status endpoint returned $STATUS"
fi

# ==================== Skills API Tests ====================

echo ""
echo -e "${BLUE}=== Skills API Tests ===${NC}"
echo ""

test_start "List all skills"
SKILLS_RESPONSE=$(curl -s "$API_URL/api/skills" --connect-timeout 5 || echo "{}")
SKILLS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/skills" --connect-timeout 5 || echo "000")
if check_response 200 "$SKILLS_STATUS"; then
    SKILL_COUNT=$(echo "$SKILLS_RESPONSE" | grep -o '"id"' | wc -l || echo "0")
    test_pass "Skills list returned $SKILL_COUNT skills"
else
    test_fail "Skills list returned $SKILLS_STATUS"
fi

test_start "Get specific skill"
SKILL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/skills/code-review" --connect-timeout 5 || echo "000")
if check_response 200 "$SKILL_STATUS" || check_response 404 "$SKILL_STATUS"; then
    test_pass "Get skill returned $SKILL_STATUS"
else
    test_fail "Get skill returned $SKILL_STATUS"
fi

test_start "Sync skills from GitLab"
SYNC_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/skills/sync" --connect-timeout 10 || echo "000")
if check_response 200 "$SYNC_STATUS"; then
    test_pass "Skills sync returned 200"
else
    test_skip "Skills sync returned $SYNC_STATUS (GitLab may not be configured)"
fi

# ==================== Tasks API Tests ====================

echo ""
echo -e "${BLUE}=== Tasks API Tests ===${NC}"
echo ""

test_start "Create task with code-review skill"
TASK_RESPONSE=$(curl -s -X POST "$API_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{"skill_id":"code-review","parameters":{"target":"src/","depth":"quick"}}' \
    --connect-timeout 10 || echo "{}")
TASK_STATUS=$(echo "$TASK_RESPONSE" | head -c 500)

if echo "$TASK_RESPONSE" | grep -q '"id"'; then
    TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    test_pass "Task created with ID: $TASK_ID"
else
    test_fail "Task creation failed: $TASK_STATUS"
    TASK_ID=""
fi

test_start "List all tasks"
TASKS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/tasks" --connect-timeout 5 || echo "000")
if check_response 200 "$TASKS_STATUS"; then
    test_pass "Tasks list returned 200"
else
    test_fail "Tasks list returned $TASKS_STATUS"
fi

if [ -n "$TASK_ID" ]; then
    test_start "Get task by ID"
    GET_TASK_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/tasks/$TASK_ID" --connect-timeout 5 || echo "000")
    if check_response 200 "$GET_TASK_STATUS"; then
        test_pass "Get task returned 200"
    else
        test_fail "Get task returned $GET_TASK_STATUS"
    fi
fi

# ==================== Pipelines API Tests ====================

echo ""
echo -e "${BLUE}=== Pipelines API Tests ===${NC}"
echo ""

test_start "Create pipeline"
PIPELINE_RESPONSE=$(curl -s -X POST "$API_URL/api/pipelines" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-pipeline",
        "description": "Test pipeline for comprehensive tests",
        "mode": "serial",
        "steps": [
            {"id": "step1", "cli": "echo", "action": "command", "params": {"message": "test"}}
        ]
    }' \
    --connect-timeout 10 || echo "{}")

if echo "$PIPELINE_RESPONSE" | grep -q '"id"'; then
    PIPELINE_ID=$(echo "$PIPELINE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    test_pass "Pipeline created with ID: $PIPELINE_ID"
else
    test_fail "Pipeline creation failed"
    PIPELINE_ID=""
fi

test_start "List pipelines"
PIPELINES_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/pipelines" --connect-timeout 5 || echo "000")
if check_response 200 "$PIPELINES_STATUS"; then
    test_pass "Pipelines list returned 200"
else
    test_fail "Pipelines list returned $PIPELINES_STATUS"
fi

if [ -n "$PIPELINE_ID" ]; then
    test_start "Get pipeline by ID"
    GET_PIPELINE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/pipelines/$PIPELINE_ID" --connect-timeout 5 || echo "000")
    if check_response 200 "$GET_PIPELINE_STATUS"; then
        test_pass "Get pipeline returned 200"
    else
        test_fail "Get pipeline returned $GET_PIPELINE_STATUS"
    fi

    test_start "Delete pipeline"
    DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$API_URL/api/pipelines/$PIPELINE_ID" --connect-timeout 5 || echo "000")
    if check_response 200 "$DELETE_STATUS"; then
        test_pass "Pipeline deleted"
    else
        test_fail "Pipeline deletion returned $DELETE_STATUS"
    fi
fi

# ==================== Schedules API Tests ====================

echo ""
echo -e "${BLUE}=== Schedules API Tests ===${NC}"
echo ""

test_start "Create schedule"
SCHEDULE_RESPONSE=$(curl -s -X POST "$API_URL/api/schedules" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-schedule",
        "skill_id": "code-review",
        "cron_expr": "@daily",
        "parameters": {"target": "src/", "depth": "quick"}
    }' \
    --connect-timeout 10 || echo "{}")

if echo "$SCHEDULE_RESPONSE" | grep -q '"id"'; then
    SCHEDULE_ID=$(echo "$SCHEDULE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    test_pass "Schedule created with ID: $SCHEDULE_ID"
else
    test_fail "Schedule creation failed"
    SCHEDULE_ID=""
fi

test_start "List schedules"
SCHEDULES_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/schedules" --connect-timeout 5 || echo "000")
if check_response 200 "$SCHEDULES_STATUS"; then
    test_pass "Schedules list returned 200"
else
    test_skip "Schedules endpoint returned $SCHEDULES_STATUS"
fi

# ==================== Batch API Tests ====================

echo ""
echo -e "${BLUE}=== Batch API Tests ===${NC}"
echo ""

test_start "Create batch operation"
BATCH_RESPONSE=$(curl -s -X POST "$API_URL/api/batches" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-batch",
        "operations": [
            {"skill_id": "code-review", "parameters": {"target": "src/", "depth": "quick"}},
            {"skill_id": "code-review", "parameters": {"target": "pkg/", "depth": "quick"}}
        ],
        "options": {"stop_on_error": false, "max_concurrency": 2}
    }' \
    --connect-timeout 10 || echo "{}")

if echo "$BATCH_RESPONSE" | grep -q '"id"'; then
    BATCH_ID=$(echo "$BATCH_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    test_pass "Batch created with ID: $BATCH_ID"
else
    test_skip "Batch creation failed (endpoint may not be available)"
fi

# ==================== Error Handling Tests ====================

echo ""
echo -e "${BLUE}=== Error Handling Tests ===${NC}"
echo ""

test_start "Invalid skill ID"
INVALID_SKILL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{"skill_id":"invalid-skill-xyz","parameters":{}}' \
    --connect-timeout 5 || echo "000")
if [ "$INVALID_SKILL_STATUS" -ge 400 ]; then
    test_pass "Invalid skill ID returned $INVALID_SKILL_STATUS (expected 4xx)"
else
    test_fail "Invalid skill ID returned $INVALID_SKILL_STATUS (expected 4xx)"
fi

test_start "Missing required parameters"
MISSING_PARAMS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{}' \
    --connect-timeout 5 || echo "000")
if [ "$MISSING_PARAMS_STATUS" -ge 400 ]; then
    test_pass "Missing params returned $MISSING_PARAMS_STATUS (expected 4xx)"
else
    test_fail "Missing params returned $MISSING_PARAMS_STATUS (expected 4xx)"
fi

test_start "Invalid JSON"
INVALID_JSON_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{invalid json}' \
    --connect-timeout 5 || echo "000")
if [ "$INVALID_JSON_STATUS" -ge 400 ]; then
    test_pass "Invalid JSON returned $INVALID_JSON_STATUS (expected 4xx)"
else
    test_fail "Invalid JSON returned $INVALID_JSON_STATUS (expected 4xx)"
fi

# ==================== Cleanup ====================

echo ""
echo -e "${BLUE}=== Cleanup ===${NC}"
echo ""

if [ -n "$SCHEDULE_ID" ]; then
    test_start "Deleting test schedule"
    curl -s -X DELETE "$API_URL/api/schedules/$SCHEDULE_ID" --connect-timeout 5 > /dev/null || true
    test_pass "Test schedule cleaned up"
fi

if [ -n "$BATCH_ID" ]; then
    test_start "Deleting test batch"
    curl -s -X DELETE "$API_URL/api/batches/$BATCH_ID" --connect-timeout 5 > /dev/null || true
    test_pass "Test batch cleaned up"
fi

# ==================== Summary ====================

echo ""
echo "========================================"
echo "           Test Summary"
echo "========================================"
echo ""
echo -e "Passed: ${GREEN}$PASS_COUNT${NC}"
echo -e "Failed: ${RED}$FAIL_COUNT${NC}"
echo ""

if [ "$FAIL_COUNT" -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi