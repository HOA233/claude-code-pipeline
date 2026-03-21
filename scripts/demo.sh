#!/bin/bash

# API Demo Script
# Demonstrates all API endpoints

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Claude Pipeline API Demo ==="
echo "Base URL: $BASE_URL"
echo ""

# 1. Check service status
echo "1. Checking service status..."
curl -s "$BASE_URL/api/status" | jq .
echo ""

# 2. List available skills
echo "2. Listing available skills..."
curl -s "$BASE_URL/api/skills" | jq .
echo ""

# 3. Create a code-review task
echo "3. Creating code-review task..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{
        "skill_id": "code-review",
        "parameters": {
            "target": "src/",
            "depth": "deep"
        }
    }')
echo "$RESPONSE" | jq .
TASK_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Task ID: $TASK_ID"
echo ""

# 4. Get task details
echo "4. Getting task details..."
curl -s "$BASE_URL/api/tasks/$TASK_ID" | jq .
echo ""

# 5. Create a test-gen task
echo "5. Creating test-gen task..."
curl -s -X POST "$BASE_URL/api/tasks" \
    -H "Content-Type: application/json" \
    -d '{
        "skill_id": "test-gen",
        "parameters": {
            "source": "src/utils/",
            "framework": "jest"
        }
    }' | jq .
echo ""

# 6. List all tasks
echo "6. Listing all tasks..."
curl -s "$BASE_URL/api/tasks" | jq .
echo ""

# 7. Sync skills
echo "7. Syncing skills from GitLab..."
curl -s -X POST "$BASE_URL/api/skills/sync" | jq .
echo ""

echo "=== Demo completed ==="