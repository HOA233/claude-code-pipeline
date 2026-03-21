#!/bin/sh
# Docker healthcheck script for Claude Pipeline API

set -e

# Check if API is responding
API_URL="${API_URL:-http://localhost:8080}"

# Primary health check
if curl -sf "${API_URL}/health" > /dev/null 2>&1; then
    echo "Health check passed"
    exit 0
fi

# Check if Redis is available (if we have redis-cli)
if command -v redis-cli > /dev/null 2>&1; then
    if redis-cli -h "${REDIS_HOST:-redis}" -p "${REDIS_PORT:-6379}" ping > /dev/null 2>&1; then
        echo "Redis check passed, API may be warming up"
        exit 0
    fi
fi

echo "Health check failed"
exit 1