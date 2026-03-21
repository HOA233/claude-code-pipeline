#!/bin/bash

# Claude Pipeline Test Script
# This script runs all tests and validates the API

set -e

echo "=== Claude Pipeline Test Script ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Redis is running
echo -e "${YELLOW}Checking Redis...${NC}"
if redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is running${NC}"
else
    echo -e "${RED}✗ Redis is not running. Please start Redis first.${NC}"
    exit 1
fi

# Run Go tests
echo ""
echo -e "${YELLOW}Running Go tests...${NC}"
go test -v ./tests/... -count=1

# Build the binary
echo ""
echo -e "${YELLOW}Building binary...${NC}"
go build -o bin/server ./cmd/server
echo -e "${GREEN}✓ Build successful${NC}"

# Start the server in background
echo ""
echo -e "${YELLOW}Starting server...${NC}"
./bin/server &
SERVER_PID=$!
sleep 2

# Function to cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    kill $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test API endpoints
echo ""
echo -e "${YELLOW}Testing API endpoints...${NC}"

BASE_URL="http://localhost:8080"

# Test health endpoint
echo ""
echo "Testing /health..."
RESPONSE=$(curl -s "$BASE_URL/health")
if echo "$RESPONSE" | grep -q "ok"; then
    echo -e "${GREEN}✓ /health endpoint working${NC}"
else
    echo -e "${RED}✗ /health endpoint failed${NC}"
fi

# Test status endpoint
echo ""
echo "Testing /api/status..."
RESPONSE=$(curl -s "$BASE_URL/api/status")
if echo "$RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✓ /api/status endpoint working${NC}"
else
    echo -e "${RED}✗ /api/status endpoint failed${NC}"
fi

# Test list skills
echo ""
echo "Testing /api/skills..."
RESPONSE=$(curl -s "$BASE_URL/api/skills")
if echo "$RESPONSE" | grep -q "code-review"; then
    echo -e "${GREEN}✓ /api/skills endpoint working${NC}"
else
    echo -e "${RED}✗ /api/skills endpoint failed${NC}"
fi

# Test create task
echo ""
echo "Testing /api/tasks (POST)..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{"skill_id":"code-review","parameters":{"target":"src/","depth":"standard"}}')
if echo "$RESPONSE" | grep -q "task-"; then
    echo -e "${GREEN}✓ /api/tasks endpoint working${NC}"
    TASK_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "  Created task: $TASK_ID"
else
    echo -e "${RED}✗ /api/tasks endpoint failed${NC}"
    echo "  Response: $RESPONSE"
fi

# Test get task
if [ ! -z "$TASK_ID" ]; then
    echo ""
    echo "Testing /api/tasks/$TASK_ID..."
    RESPONSE=$(curl -s "$BASE_URL/api/tasks/$TASK_ID")
    if echo "$RESPONSE" | grep -q "$TASK_ID"; then
        echo -e "${GREEN}✓ /api/tasks/{id} endpoint working${NC}"
    else
        echo -e "${RED}✗ /api/tasks/{id} endpoint failed${NC}"
    fi
fi

echo ""
echo -e "${GREEN}=== All tests completed ===${NC}"