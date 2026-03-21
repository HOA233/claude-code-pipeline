#!/bin/bash

# Setup script for Claude Pipeline
# Installs dependencies and prepares the environment

set -e

echo "========================================"
echo "Claude Pipeline Setup"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check dependencies
check_command() {
    local cmd="$1"
    local name="$2"
    local install_cmd="$3"

    if command -v "$cmd" &> /dev/null; then
        echo -e "${GREEN}✓ $name is installed${NC}"
        return 0
    else
        echo -e "${YELLOW}✗ $name is not installed${NC}"
        if [ -n "$install_cmd" ]; then
            echo "  Install with: $install_cmd"
        fi
        return 1
    fi
}

echo "Checking dependencies..."
echo ""

# Required dependencies
check_command "go" "Go" "https://go.dev/doc/install"
check_command "redis-cli" "Redis CLI" "apt install redis-tools / brew install redis"
check_command "docker" "Docker" "https://docs.docker.com/get-docker/"
check_command "docker-compose" "Docker Compose" "https://docs.docker.com/compose/install/"

echo ""
echo "Checking optional dependencies..."

# Optional dependencies
check_command "kubectl" "kubectl" "https://kubernetes.io/docs/tasks/tools/"
check_command "helm" "Helm" "https://helm.sh/docs/intro/install/"

echo ""
echo "Setting up environment..."

# Copy .env.example to .env if not exists
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file${NC}"
    echo -e "${YELLOW}  Please edit .env with your API keys${NC}"
else
    echo -e "${GREEN}✓ .env file exists${NC}"
fi

# Create logs directory
mkdir -p logs
echo -e "${GREEN}✓ Created logs directory${NC}"

# Download Go dependencies
echo ""
echo "Downloading Go dependencies..."
if command -v go &> /dev/null; then
    go mod download
    go mod tidy
    echo -e "${GREEN}✓ Go dependencies installed${NC}"
fi

# Install frontend dependencies
echo ""
echo "Installing frontend dependencies..."
if [ -d "frontend" ] && command -v npm &> /dev/null; then
    cd frontend
    npm install
    cd ..
    echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
fi

echo ""
echo "========================================"
echo "Setup Complete!"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your API keys"
echo "2. Start Redis: docker-compose up -d redis"
echo "3. Start the server: make run"
echo "4. Run tests: ./scripts/comprehensive_test.sh"
echo ""