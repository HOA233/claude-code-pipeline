#!/bin/bash

# Development Setup Script for Claude CLI Pipeline Service
# Usage: ./scripts/dev.sh [command]
#
# This script runs everything locally without Docker.
# Prerequisites: Go 1.22+, Redis (installed locally), Node.js 18+ (for frontend)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# PID files
PID_DIR="/tmp/claude-pipeline"
API_PID_FILE="$PID_DIR/api.pid"
FRONTEND_PID_FILE="$PID_DIR/frontend.pid"
REDIS_PID_FILE="$PID_DIR/redis.pid"

# Default ports
REDIS_PORT=${REDIS_PORT:-6379}
API_PORT=${API_PORT:-8080}
FRONTEND_PORT=${FRONTEND_PORT:-3000}

# Ensure PID directory exists
mkdir -p "$PID_DIR"
mkdir -p logs

# Print banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║         Claude CLI Pipeline - Development Script          ║"
    echo "║                  (Local Environment)                       ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Print usage
print_usage() {
    echo -e "${YELLOW}Usage:${NC} ./scripts/dev.sh [command]"
    echo ""
    echo -e "${GREEN}Commands:${NC}"
    echo "  setup       - Initial project setup (install dependencies)"
    echo "  dev         - Start all services locally (Ctrl+C to stop)"
    echo "  start       - Start all services in background"
    echo "  stop        - Stop all services"
    echo "  restart     - Restart all services"
    echo "  status      - Show status of services"
    echo "  logs        - Show API logs"
    echo "  test        - Run all tests"
    echo "  lint        - Run linter"
    echo "  fmt         - Format code"
    echo "  clean       - Clean build artifacts"
    echo "  build       - Build binary"
    echo "  check       - Run all checks (fmt, lint, test)"
    echo "  help        - Show this help message"
    echo ""
    echo -e "${YELLOW}Prerequisites:${NC}"
    echo "  - Go 1.22+"
    echo "  - Redis (running locally on port $REDIS_PORT)"
    echo "  - Node.js 18+ (for frontend)"
    echo ""
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if port is in use
port_in_use() {
    local port=$1
    if command_exists lsof; then
        lsof -i :$port >/dev/null 2>&1
    elif command_exists netstat; then
        netstat -an 2>/dev/null | grep -q ":$port .*LISTEN"
    elif command_exists ss; then
        ss -ln 2>/dev/null | grep -q ":$port"
    else
        return 1
    fi
}

# Wait for port to be available
wait_for_port() {
    local port=$1
    local service=$2
    local max_wait=${3:-30}

    echo -e "${YELLOW}Waiting for $service on port $port...${NC}"

    for i in $(seq 1 $max_wait); do
        if port_in_use $port; then
            echo -e "${GREEN}✓ $service is ready${NC}"
            return 0
        fi
        sleep 1
    done

    echo -e "${RED}$service failed to start on port $port${NC}"
    return 1
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}Checking prerequisites...${NC}"

    local missing=()

    if ! command_exists go; then
        missing+=("Go 1.22+")
    fi

    if ! command_exists redis-server && ! command_exists redis-cli; then
        missing+=("Redis")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "${RED}Missing prerequisites:${NC}"
        for dep in "${missing[@]}"; do
            echo "  - $dep"
        done
        echo ""
        echo "Please install missing dependencies and try again."
        exit 1
    fi

    echo -e "${GREEN}✓ All prerequisites installed${NC}"
}

# Check Redis connection
check_redis() {
    if redis-cli -p $REDIS_PORT ping 2>/dev/null | grep -q "PONG"; then
        return 0
    else
        return 1
    fi
}

# Start Redis locally
start_redis() {
    echo -e "${BLUE}Starting Redis...${NC}"

    # Check if Redis is already running
    if check_redis; then
        echo -e "${GREEN}✓ Redis already running on port $REDIS_PORT${NC}"
        return 0
    fi

    # Try to start Redis
    if command_exists redis-server; then
        # Start Redis in background
        redis-server --port $REDIS_PORT --daemonize yes --pidfile "$REDIS_PID_FILE" 2>/dev/null

        sleep 2

        # Verify Redis started
        if check_redis; then
            echo -e "${GREEN}✓ Redis started on port $REDIS_PORT${NC}"
            return 0
        fi
    fi

    echo -e "${RED}Failed to start Redis${NC}"
    echo -e "${YELLOW}Please start Redis manually: redis-server${NC}"
    return 1
}

# Stop all services
stop_all() {
    echo -e "${BLUE}Stopping all services...${NC}"

    # Stop API process
    if [ -f "$API_PID_FILE" ]; then
        local pid=$(cat "$API_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${YELLOW}Stopping API (PID: $pid)...${NC}"
            kill "$pid" 2>/dev/null || true
            for i in {1..10}; do
                if ! kill -0 "$pid" 2>/dev/null; then
                    break
                fi
                sleep 1
            done
            kill -9 "$pid" 2>/dev/null || true
        fi
        rm -f "$API_PID_FILE"
    fi

    # Stop Frontend process
    if [ -f "$FRONTEND_PID_FILE" ]; then
        local pid=$(cat "$FRONTEND_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${YELLOW}Stopping Frontend (PID: $pid)...${NC}"
            kill "$pid" 2>/dev/null || true
            kill -9 "$pid" 2>/dev/null || true
        fi
        rm -f "$FRONTEND_PID_FILE"
    fi

    # Stop Redis if we started it
    if [ -f "$REDIS_PID_FILE" ]; then
        local pid=$(cat "$REDIS_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${YELLOW}Stopping Redis (PID: $pid)...${NC}"
            redis-cli -p $REDIS_PORT shutdown 2>/dev/null || kill "$pid" 2>/dev/null || true
        fi
        rm -f "$REDIS_PID_FILE"
    fi

    echo -e "${GREEN}✓ All services stopped${NC}"
}

# Cleanup function for trap
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down...${NC}"
    stop_all
    exit 0
}

# Start all services in foreground (with cleanup on exit)
start_all_foreground() {
    echo -e "${BLUE}Starting all services locally...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
    echo ""

    # Set trap for cleanup
    trap cleanup SIGINT SIGTERM

    # Start Redis
    start_redis || exit 1

    # Load environment
    if [ -f .env ]; then
        echo -e "${YELLOW}Loading .env file...${NC}"
        export $(cat .env | grep -v '^#' | xargs 2>/dev/null)
    fi

    # Start API in background
    echo -e "${BLUE}Starting API server...${NC}"
    go run ./cmd/server > logs/api.log 2>&1 &
    API_PID=$!
    echo $API_PID > "$API_PID_FILE"
    echo -e "${GREEN}✓ API started (PID: $API_PID)${NC}"

    sleep 2

    # Check API is running
    if ! kill -0 $API_PID 2>/dev/null; then
        echo -e "${RED}API failed to start. Check logs/api.log${NC}"
        cat logs/api.log
        exit 1
    fi

    # Start Frontend in background (if exists)
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo -e "${BLUE}Starting Frontend...${NC}"
        cd frontend

        # Check if node_modules exists
        if [ ! -d "node_modules" ]; then
            echo -e "${YELLOW}Installing frontend dependencies...${NC}"
            npm install
        fi

        npm run dev > ../logs/frontend.log 2>&1 &
        FRONTEND_PID=$!
        echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
        cd ..
        echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
        sleep 3
    fi

    # Show status
    echo ""
    echo -e "${GREEN}══════════════════════════════════════${NC}"
    echo -e "${GREEN}  All services running!${NC}"
    echo -e "${GREEN}══════════════════════════════════════${NC}"
    echo ""
    echo "Services:"
    echo "  - Redis:     localhost:$REDIS_PORT"
    echo "  - API:       http://localhost:$API_PORT"
    echo "  - Health:    http://localhost:$API_PORT/health"
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo "  - Frontend:  http://localhost:$FRONTEND_PORT"
    fi
    echo ""
    echo "Logs:"
    echo "  - API:       logs/api.log"
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo "  - Frontend:  logs/frontend.log"
    fi
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
    echo ""

    # Wait for processes
    wait $API_PID $FRONTEND_PID 2>/dev/null
}

# Start all services in background
start_all_background() {
    echo -e "${BLUE}Starting all services in background...${NC}"

    # Start Redis
    start_redis || exit 1

    # Load environment
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs 2>/dev/null)
    fi

    # Start API
    echo -e "${BLUE}Starting API server...${NC}"
    go run ./cmd/server > logs/api.log 2>&1 &
    API_PID=$!
    echo $API_PID > "$API_PID_FILE"

    sleep 2

    # Start Frontend
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo -e "${BLUE}Starting Frontend...${NC}"
        cd frontend
        npm run dev > ../logs/frontend.log 2>&1 &
        FRONTEND_PID=$!
        echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
        cd ..
    fi

    echo -e "${GREEN}✓ All services started${NC}"
    echo ""
    echo "Services:"
    echo "  - Redis:     localhost:$REDIS_PORT"
    echo "  - API:       http://localhost:$API_PORT"
    echo "  - Health:    http://localhost:$API_PORT/health"
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo "  - Frontend:  http://localhost:$FRONTEND_PORT"
    fi
    echo ""
    echo "To stop: ./scripts/dev.sh stop"
    echo "To view logs: tail -f logs/api.log"
}

# Restart all services
restart_all() {
    stop_all
    sleep 2
    start_all_background
}

# Show status
show_status() {
    echo -e "${BLUE}Service Status:${NC}"
    echo ""

    # Check Redis
    if check_redis; then
        echo -e "Redis:   ${GREEN}Running${NC} (localhost:$REDIS_PORT)"
    else
        echo -e "Redis:   ${RED}Stopped${NC}"
    fi

    # Check API
    if [ -f "$API_PID_FILE" ]; then
        local pid=$(cat "$API_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "API:     ${GREEN}Running${NC} (PID: $pid, http://localhost:$API_PORT)"
        else
            echo -e "API:     ${RED}Stopped${NC}"
            rm -f "$API_PID_FILE"
        fi
    else
        if curl -s http://localhost:$API_PORT/health > /dev/null 2>&1; then
            echo -e "API:     ${GREEN}Running${NC} (http://localhost:$API_PORT)"
        else
            echo -e "API:     ${RED}Stopped${NC}"
        fi
    fi

    # Check Frontend
    if [ -f "$FRONTEND_PID_FILE" ]; then
        local pid=$(cat "$FRONTEND_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "Frontend: ${GREEN}Running${NC} (PID: $pid, http://localhost:$FRONTEND_PORT)"
        else
            echo -e "Frontend: ${RED}Stopped${NC}"
            rm -f "$FRONTEND_PID_FILE"
        fi
    else
        if curl -s http://localhost:$FRONTEND_PORT > /dev/null 2>&1; then
            echo -e "Frontend: ${GREEN}Running${NC} (http://localhost:$FRONTEND_PORT)"
        else
            echo -e "Frontend: ${RED}Stopped${NC}"
        fi
    fi
}

# Show logs
show_logs() {
    local log_file=${1:-api}
    tail -f "logs/${log_file}.log"
}

# Run tests
run_tests() {
    echo -e "${BLUE}Running tests...${NC}"
    go test -v ./tests/...
}

# Run linter
run_lint() {
    echo -e "${BLUE}Running linter...${NC}"

    if command_exists golangci-lint; then
        golangci-lint run ./...
    else
        echo -e "${YELLOW}golangci-lint not installed, using go vet...${NC}"
        go vet ./...
    fi
}

# Format code
format_code() {
    echo -e "${BLUE}Formatting code...${NC}"
    go fmt ./...
    echo -e "${GREEN}✓ Code formatted${NC}"
}

# Clean build artifacts
clean_artifacts() {
    echo -e "${BLUE}Cleaning build artifacts...${NC}"
    rm -rf bin/
    rm -f coverage.out coverage.html
    rm -rf "$PID_DIR"
    rm -rf logs/
    echo -e "${GREEN}✓ Cleaned${NC}"
}

# Build binary
build_binary() {
    echo -e "${BLUE}Building binary...${NC}"
    mkdir -p bin
    go build -o bin/server ./cmd/server
    go build -o bin/cli ./cmd/cli
    echo -e "${GREEN}✓ Binaries built in bin/${NC}"
}

# Setup project
setup_project() {
    echo -e "${BLUE}Setting up project...${NC}"

    check_prerequisites

    # Create .env if not exists
    if [ ! -f .env ]; then
        echo -e "${YELLOW}Creating .env from .env.example...${NC}"
        cp .env.example .env
    fi

    # Install Go dependencies
    echo -e "${YELLOW}Installing Go dependencies...${NC}"
    go mod download
    go mod tidy

    # Install frontend dependencies
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo -e "${YELLOW}Installing frontend dependencies...${NC}"
        cd frontend
        npm install
        cd ..
    fi

    # Make scripts executable
    chmod +x scripts/*.sh 2>/dev/null || true

    echo -e "${GREEN}✓ Setup complete!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Edit .env with your configuration"
    echo "  2. Make sure Redis is installed and running"
    echo "  3. Run: ./scripts/dev.sh dev"
}

# Run all checks
run_checks() {
    echo -e "${BLUE}Running all checks...${NC}"
    echo ""

    echo -e "${YELLOW}1. Formatting...${NC}"
    format_code

    echo -e "${YELLOW}2. Linting...${NC}"
    run_lint

    echo -e "${YELLOW}3. Testing...${NC}"
    run_tests

    echo ""
    echo -e "${GREEN}✓ All checks passed${NC}"
}

# Main
print_banner

case "${1:-help}" in
    setup)
        setup_project
        ;;
    dev)
        start_all_foreground
        ;;
    start)
        start_all_background
        ;;
    stop)
        stop_all
        ;;
    restart)
        restart_all
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    test)
        run_tests
        ;;
    lint)
        run_lint
        ;;
    fmt)
        format_code
        ;;
    clean)
        clean_artifacts
        ;;
    build)
        build_binary
        ;;
    check)
        run_checks
        ;;
    help|--help|-h)
        print_usage
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        print_usage
        exit 1
        ;;
esac