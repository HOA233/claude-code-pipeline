#!/bin/bash

# Clean up script for Claude Pipeline
# Removes generated files, logs, and temporary data

set -e

echo "========================================"
echo "Claude Pipeline Cleanup"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Ask for confirmation
read -p "This will remove all generated files, logs, and build artifacts. Continue? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Clean Go build artifacts
echo -e "${YELLOW}Cleaning Go build artifacts...${NC}"
rm -rf bin/
rm -f coverage.out coverage.html
echo -e "${GREEN}✓ Go artifacts cleaned${NC}"

# Clean logs
echo -e "${YELLOW}Cleaning logs...${NC}"
rm -rf logs/*
echo -e "${GREEN}✓ Logs cleaned${NC}"

# Clean frontend build
echo -e "${YELLOW}Cleaning frontend build...${NC}"
if [ -d "frontend" ]; then
    rm -rf frontend/dist
    rm -rf frontend/node_modules
    echo -e "${GREEN}✓ Frontend cleaned${NC}"
fi

# Clean Docker
echo -e "${YELLOW}Cleaning Docker resources...${NC}"
read -p "Remove Docker containers and volumes? (y/N) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose down -v --remove-orphans 2>/dev/null || true
    echo -e "${GREEN}✓ Docker resources cleaned${NC}"
fi

# Clean temporary files
echo -e "${YELLOW}Cleaning temporary files...${NC}"
find . -type f -name "*.tmp" -delete 2>/dev/null || true
find . -type f -name "*.log" -delete 2>/dev/null || true
echo -e "${GREEN}✓ Temporary files cleaned${NC}"

echo ""
echo "========================================"
echo "Cleanup Complete!"
echo "========================================"