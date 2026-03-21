#!/bin/bash
# WebSocket Stress Test for Claude Pipeline API

set -e

BASE_URL="${BASE_URL:-ws://localhost:8080}"
CONNECTIONS="${CONNECTIONS:-50}"
DURATION="${DURATION:-30}"
MESSAGE_RATE="${MESSAGE_RATE:-10}"

echo "============================================"
echo "WebSocket Stress Test"
echo "============================================"
echo "Base URL: $BASE_URL"
echo "Connections: $CONNECTIONS"
echo "Duration: ${DURATION}s"
echo "Message Rate: ${MESSAGE_RATE}/s per connection"
echo ""

# Check for websocat
if ! command -v websocat &> /dev/null; then
    echo "Installing websocat..."
    cargo install websocat 2>/dev/null || {
        echo "websocat not available, using wscat instead"
        npm install -g wscat
    }
fi

# Statistics
TOTAL_MESSAGES=0
TOTAL_ERRORS=0
CONNECTION_ERRORS=0

# Test basic WebSocket connection
test_basic_connection() {
    echo "=== Testing Basic WebSocket Connection ==="

    # Test global WebSocket
    if timeout 5 websocat "$BASE_URL/ws" <<< '{"type":"ping"}' 2>/dev/null; then
        echo "✓ Global WebSocket connection successful"
    else
        echo "✗ Global WebSocket connection failed"
    fi

    # Test SSE endpoint
    if timeout 5 curl -s "$BASE_URL/sse" | head -1 | grep -q .; then
        echo "✓ SSE endpoint responding"
    else
        echo "✗ SSE endpoint not responding"
    fi
}

# Run concurrent WebSocket connections
run_stress_test() {
    echo ""
    echo "=== Running Stress Test ==="

    local pids=()
    local start_time=$(date +%s)

    for i in $(seq 1 $CONNECTIONS); do
        (
            local msg_count=0
            local err_count=0
            local end_time=$((start_time + DURATION))

            while [ $(date +%s) -lt $end_time ]; do
                # Send message and check response
                if websocat -n1 "$BASE_URL/ws" <<< '{"type":"ping","timestamp":"'$(date -Iseconds)'"}' 2>/dev/null; then
                    ((msg_count++))
                else
                    ((err_count++))
                fi

                sleep $((1 / MESSAGE_RATE))
            done

            echo "Connection $i: $msg_count messages, $err_count errors"
        ) &
        pids+=($!)
    done

    # Wait for all connections
    echo "Waiting for test to complete..."
    for pid in "${pids[@]}"; do
        wait $pid
    done
}

# Test subscription endpoints
test_subscriptions() {
    echo ""
    echo "=== Testing Subscription Endpoints ==="

    # Create a test task first
    TASK_RESPONSE=$(curl -s -X POST "$BASE_URL/api/tasks" \
        -H "Content-Type: application/json" \
        -d '{"skill_id":"code-review","parameters":{"target":"src/"}}')

    TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$TASK_ID" ]; then
        echo "Created test task: $TASK_ID"

        # Test task WebSocket subscription
        echo "Testing task WebSocket subscription..."
        timeout 10 websocat "$BASE_URL/ws/tasks/$TASK_ID" 2>/dev/null &
        WS_PID=$!

        sleep 2
        kill $WS_PID 2>/dev/null || true

        echo "✓ Task subscription test completed"
    else
        echo "✗ Failed to create test task"
    fi
}

# Run benchmarks
run_benchmarks() {
    echo ""
    echo "=== Running Benchmarks ==="

    # Connection establishment time
    echo "Measuring connection establishment time..."
    for i in {1..10}; do
        START=$(date +%s%N)
        websocat -n1 "$BASE_URL/ws" <<< '{"type":"ping"}' >/dev/null 2>&1
        END=$(date +%s%N)
        DURATION_MS=$(( (END - START) / 1000000 ))
        echo "  Connection $i: ${DURATION_MS}ms"
    done

    # Message throughput
    echo ""
    echo "Measuring message throughput..."
    MESSAGES_SENT=0
    START_TIME=$(date +%s)

    for i in {1..100}; do
        websocat -n1 "$BASE_URL/ws" <<< '{"type":"ping","seq":'$i'}' >/dev/null 2>&1 && ((MESSAGES_SENT++))
    done

    END_TIME=$(date +%s)
    ELAPSED=$((END_TIME - START_TIME))
    THROUGHPUT=$((MESSAGES_SENT / ELAPSED))

    echo "  Messages sent: $MESSAGES_SENT"
    echo "  Elapsed time: ${ELAPSED}s"
    echo "  Throughput: ${THROUGHPUT} msg/s"
}

# Main execution
echo ""
test_basic_connection
run_stress_test
test_subscriptions
run_benchmarks

echo ""
echo "============================================"
echo "WebSocket Stress Test Complete"
echo "============================================"