#!/bin/bash

# Development Setup Script for Claude CLI Pipeline Service
# Usage: ./scripts/dev.sh [command]

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Docker Compose command (supports both v1 and v2)
if docker compose version &>/dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo -e "${RED}Error: Docker Compose not found${NC}"
    exit 1
fi

# PID files
PID_DIR="/tmp/claude-pipeline"
API_PID_FILE="$PID_DIR/api.pid"
FRONTEND_PID_FILE="$PID_DIR/frontend.pid"

# Ensure PID directory exists
mkdir -p "$PID_DIR"

# Print banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║         Claude CLI Pipeline - Development Script          ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Print usage
print_usage() {
    echo -e "${YELLOW}Usage:${NC} ./scripts/dev.sh [command]"
    echo ""
    echo -e "${GREEN}Commands:${NC}"
    echo "  setup       - Initial project setup (install dependencies)"
    echo "  start       - Start all services (Redis + API)"
    echo "  stop        - Stop all services"
    echo "  restart     - Restart all services"
    echo "  status      - Show status of services"
    echo "  logs        - Show logs from all services"
    echo "  test        - Run all tests"
    echo "  test-watch  - Run tests in watch mode"
    echo "  lint        - Run linter"
    echo "  fmt         - Format code"
    echo "  clean       - Clean build artifacts"
    echo "  db          - Start only Redis"
    echo "  db-cli      - Open Redis CLI"
    echo "  db-reset    - Reset Redis data"
    echo "  api         - Start only API server (foreground)"
    echo "  api-bg      - Start API server in background"
    echo "  build       - Build binary"
    echo "  frontend    - Start frontend dev server"
    echo "  install     - Install all dependencies"
    echo "  check       - Run all checks (fmt, lint, test)"
    echo "  dev         - Start all services in foreground (Ctrl+C to stop all)"
    echo "  help        - Show this help message"
    echo ""
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}Checking prerequisites...${NC}"

    local missing=()

    if ! command_exists go; then
        missing+=("Go 1.22+")
    fi

    if ! command_exists docker; then
        missing+=("Docker")
    fi

    if ! docker compose version &>/dev/null && ! command_exists docker-compose; then
        missing+=("Docker Compose")
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
    echo "  2. Run: ./scripts/dev.sh dev"
}

# Start Redis
start_db() {
    echo -e "${BLUE}Starting Redis...${NC}"

    # Check if Docker is running
    if ! docker info &>/dev/null; then
        echo -e "${RED}Error: Docker is not running. Please start Docker first.${NC}"
        exit 1
    fi

    # Start Redis container
    $DOCKER_COMPOSE up -d redis

    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to start Redis container${NC}"
        exit 1
    fi

    sleep 2

    # Wait for Redis to be ready
    echo -e "${YELLOW}Waiting for Redis...${NC}"
    for i in {1..30}; do
        if $DOCKER_COMPOSE exec -T redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
            echo -e "${GREEN}✓ Redis is ready${NC}"
            return 0
        fi
        sleep 1
    done

    echo -e "${RED}Redis failed to start${NC}"
    echo -e "${YELLOW}Try running: $DOCKER_COMPOSE logs redis${NC}"
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
            # Wait for process to die
            for i in {1..10}; do
                if ! kill -0 "$pid" 2>/dev/null; then
                    break
                fi
                sleep 1
            done
            # Force kill if still running
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

    # Stop Docker containers
    $DOCKER_COMPOSE down 2>/dev/null || true

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
    echo -e "${BLUE}Starting all services in foreground...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
    echo ""

    # Set trap for cleanup
    trap cleanup SIGINT SIGTERM

    # Start Redis
    start_db

    # Create logs directory
    mkdir -p logs

    # Start API in background
    echo -e "${BLUE}Starting API server...${NC}"
    go run ./cmd/server &
    API_PID=$!
    echo $API_PID > "$API_PID_FILE"
    echo -e "${GREEN}✓ API started (PID: $API_PID)${NC}"

    # Wait a moment for API to start
    sleep 2

    # Start Frontend in background (if exists)
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo -e "${BLUE}Starting Frontend...${NC}"
        cd frontend
        npm run dev &
        FRONTEND_PID=$!
        echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
        cd ..
        echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
        sleep 2
    fi

    # Show status
    echo ""
    echo -e "${GREEN}══════════════════════════════════════${NC}"
    echo -e "${GREEN}  All services running!${NC}"
    echo -e "${GREEN}══════════════════════════════════════${NC}"
    echo ""
    echo "Services:"
    echo "  - Redis:     localhost:6379"
    echo "  - API:       http://localhost:8080"
    echo "  - Health:    http://localhost:8080/health"
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo "  - Frontend:  http://localhost:3000"
    fi
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
    echo ""

    # Wait for processes
    wait $API_PID $FRONTEND_PID 2>/dev/null
}

# Start all services in background
start_all_background() {
    echo -e "${BLUE}Starting all services...${NC}"

    # Start Redis
    start_db

    # Start API in background
    echo -e "${BLUE}Starting API server...${NC}"
    nohup go run ./cmd/server > logs/api.log 2>&1 &
    API_PID=$!
    echo $API_PID > "$API_PID_FILE"

    sleep 2

    echo -e "${GREEN}✓ All services started${NC}"
    echo ""
    echo "Services:"
    echo "  - API:    http://localhost:8080"
    echo "  - Redis:  localhost:6379"
    echo "  - Health: http://localhost:8080/health"
    echo ""
    echo "To stop: ./scripts/dev.sh stop"
    echo "To view logs: ./scripts/dev.sh logs"
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
    if $DOCKER_COMPOSE ps redis 2>/dev/null | grep -q "Up" || $DOCKER_COMPOSE ps 2>/dev/null | grep redis | grep -q "Up"; then
        echo -e "Redis:   ${GREEN}Running${NC} (localhost:6379)"
    else
        echo -e "Redis:   ${RED}Stopped${NC}"
    fi

    # Check API
    if [ -f "$API_PID_FILE" ]; then
        local pid=$(cat "$API_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "API:     ${GREEN}Running${NC} (PID: $pid, http://localhost:8080)"
        else
            echo -e "API:     ${RED}Stopped (stale PID file)${NC}"
            rm -f "$API_PID_FILE"
        fi
    else
        # Try to check if API is responding
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "API:     ${GREEN}Running${NC} (http://localhost:8080)"
        else
            echo -e "API:     ${RED}Stopped${NC}"
        fi
    fi

    # Check Frontend
    if [ -f "$FRONTEND_PID_FILE" ]; then
        local pid=$(cat "$FRONTEND_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "Frontend: ${GREEN}Running${NC} (PID: $pid, http://localhost:3000)"
        else
            echo -e "Frontend: ${RED}Stopped${NC}"
            rm -f "$FRONTEND_PID_FILE"
        fi
    else
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            echo -e "Frontend: ${GREEN}Running${NC} (http://localhost:3000)"
        else
            echo -e "Frontend: ${RED}Stopped${NC}"
        fi
    fi
}

# Show logs
show_logs() {
    local service=${1:-}

    if [ -n "$service" ]; then
        $DOCKER_COMPOSE logs -f "$service"
    else
        $DOCKER_COMPOSE logs -f
    fi
}

# Run tests
run_tests() {
    echo -e "${BLUE}Running tests...${NC}"
    go test -v ./tests/...
}

# Run tests in watch mode
run_tests_watch() {
    echo -e "${BLUE}Running tests in watch mode...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop${NC}"

    # Install reflex if not present
    if ! command_exists reflex; then
        echo "Installing reflex for watch mode..."
        go install github.com/cespare/reflex@latest
    fi

    reflex -r '\.go$' -s -- sh -c "go test -v ./tests/..."
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
    echo -e "${GREEN}✓ Cleaned${NC}"
}

# Open Redis CLI
db_cli() {
    $DOCKER_COMPOSE exec redis redis-cli
}

# Reset Redis data
db_reset() {
    echo -e "${YELLOW}Resetting Redis data...${NC}"
    $DOCKER_COMPOSE exec -T redis redis-cli FLUSHALL
    echo -e "${GREEN}✓ Redis data reset${NC}"
}

# Start API in foreground
start_api_foreground() {
    echo -e "${BLUE}Starting API server in foreground...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
    echo ""

    # Ensure Redis is running
    start_db

    # Run API
    go run ./cmd/server
}

# Start API in background
start_api_background() {
    echo -e "${BLUE}Starting API server in background...${NC}"

    # Ensure Redis is running
    start_db

    # Create logs directory
    mkdir -p logs

    # Start API
    nohup go run ./cmd/server > logs/api.log 2>&1 &
    API_PID=$!
    echo $API_PID > "$API_PID_FILE"

    echo -e "${GREEN}✓ API started (PID: $API_PID)${NC}"
    echo "  URL: http://localhost:8080"
    echo "  Logs: logs/api.log"
    echo ""
    echo "To stop: ./scripts/dev.sh stop"
}

# Build binary
build_binary() {
    echo -e "${BLUE}Building binary...${NC}"
    mkdir -p bin
    go build -o bin/server ./cmd/server
    go build -o bin/cli ./cmd/cli
    echo -e "${GREEN}✓ Binaries built in bin/${NC}"
}

# Start frontend
start_frontend() {
    echo -e "${BLUE}Starting frontend dev server...${NC}"

    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        cd frontend
        npm run dev &
        FRONTEND_PID=$!
        echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
        cd ..
        echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
        echo "  URL: http://localhost:3000"
    else
        echo -e "${RED}Frontend directory not found${NC}"
    fi
}

# Install all dependencies
install_deps() {
    echo -e "${BLUE}Installing all dependencies...${NC}"

    # Go dependencies
    echo -e "${YELLOW}Go dependencies...${NC}"
    go mod download
    go mod tidy

    # Frontend dependencies
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        echo -e "${YELLOW}Frontend dependencies...${NC}"
        cd frontend
        npm install
        cd ..
    fi

    # Development tools
    echo -e "${YELLOW}Development tools...${NC}"
    go install github.com/cespare/reflex@latest 2>/dev/null || true

    echo -e "${GREEN}✓ All dependencies installed${NC}"
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
    start)
        start_all_background
        ;;
    dev)
        start_all_foreground
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
    test-watch)
        run_tests_watch
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
    db)
        start_db
        ;;
    db-cli)
        db_cli
        ;;
    db-reset)
        db_reset
        ;;
    api)
        start_api_foreground
        ;;
    api-bg)
        start_api_background
        ;;
    build)
        build_binary
        ;;
    frontend)
        start_frontend
        ;;
    install)
        install_deps
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