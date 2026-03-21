#!/bin/bash

# Build script for Claude Pipeline
# Builds binaries for multiple platforms

set -e

VERSION="${VERSION:-dev}"
BUILD_DIR="build"
BINARY_NAME="claude-pipeline"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================"
echo "Building Claude Pipeline"
echo "Version: $VERSION"
echo "========================================"
echo ""

# Create build directory
mkdir -p $BUILD_DIR

# Build info
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

# Build for multiple platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r OS ARCH <<< "$PLATFORM"

    echo -e "${YELLOW}Building for $OS/$ARCH...${NC}"

    OUTPUT="${BUILD_DIR}/${BINARY_NAME}-${OS}-${ARCH}"

    if [ "$OS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build \
        -ldflags "$LDFLAGS" \
        -o "$OUTPUT" \
        ./cmd/server

    echo -e "${GREEN}✓ Built $OUTPUT${NC}"
done

# Build CLI tool
echo ""
echo -e "${YELLOW}Building CLI tool...${NC}"

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r OS ARCH <<< "$PLATFORM"

    OUTPUT="${BUILD_DIR}/pipeline-cli-${OS}-${ARCH}"

    if [ "$OS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build \
        -ldflags "$LDFLAGS" \
        -o "$OUTPUT" \
        ./cmd/cli

    echo -e "${GREEN}✓ Built $OUTPUT${NC}"
done

# Build frontend
echo ""
echo -e "${YELLOW}Building frontend...${NC}"

if [ -d "frontend" ]; then
    cd frontend
    npm run build
    cd ..
    mkdir -p ${BUILD_DIR}/frontend
    cp -r frontend/dist/* ${BUILD_DIR}/frontend/
    echo -e "${GREEN}✓ Frontend built${NC}"
fi

echo ""
echo "========================================"
echo "Build Complete!"
echo "========================================"
echo ""
echo "Binaries available in: $BUILD_DIR"
ls -la $BUILD_DIR/