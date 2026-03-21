#!/bin/bash

# Deploy script for Claude Pipeline
# Supports Docker Compose and Kubernetes deployments

set -e

DEPLOY_TYPE="${1:-docker}"
VERSION="${VERSION:-latest}"
NAMESPACE="${NAMESPACE:-claude-pipeline}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "========================================"
echo "Claude Pipeline Deployment"
echo "========================================"
echo "Type: $DEPLOY_TYPE"
echo "Version: $VERSION"
echo ""

case $DEPLOY_TYPE in
    docker)
        echo -e "${BLUE}Deploying with Docker Compose...${NC}"
        echo ""

        # Build images
        echo -e "${YELLOW}Building images...${NC}"
        docker-compose build

        # Stop existing containers
        echo -e "${YELLOW}Stopping existing containers...${NC}"
        docker-compose down

        # Start containers
        echo -e "${YELLOW}Starting containers...${NC}"
        docker-compose up -d

        # Wait for health
        echo -e "${YELLOW}Waiting for service to be healthy...${NC}"
        sleep 5

        if curl -sf http://localhost:8080/health > /dev/null; then
            echo -e "${GREEN}✓ Service is healthy${NC}"
        else
            echo -e "${RED}✗ Service health check failed${NC}"
            docker-compose logs
            exit 1
        fi

        echo ""
        echo -e "${GREEN}Deployment successful!${NC}"
        echo "API: http://localhost:8080"
        echo "Frontend: http://localhost:3000"
        ;;

    docker-prod)
        echo -e "${BLUE}Deploying with Docker Compose (Production)...${NC}"
        echo ""

        export VERSION=$VERSION

        docker-compose -f docker-compose.prod.yml up -d --build

        echo -e "${GREEN}✓ Production deployment complete${NC}"
        ;;

    kubernetes)
        echo -e "${BLUE}Deploying to Kubernetes...${NC}"
        echo ""

        # Check kubectl
        if ! command -v kubectl &> /dev/null; then
            echo -e "${RED}kubectl is not installed${NC}"
            exit 1
        fi

        # Create namespace if not exists
        kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

        # Apply manifests
        echo -e "${YELLOW}Applying Kubernetes manifests...${NC}"
        kubectl apply -f deploy/kubernetes/ -n $NAMESPACE

        # Wait for rollout
        echo -e "${YELLOW}Waiting for deployment...${NC}"
        kubectl rollout status deployment/claude-pipeline-api -n $NAMESPACE --timeout=300s

        echo ""
        echo -e "${GREEN}Deployment successful!${NC}"
        kubectl get pods -n $NAMESPACE
        ;;

    helm)
        echo -e "${BLUE}Deploying with Helm...${NC}"
        echo ""

        # Check helm
        if ! command -v helm &> /dev/null; then
            echo -e "${RED}helm is not installed${NC}"
            exit 1
        fi

        # Deploy with Helm
        echo -e "${YELLOW}Deploying Helm chart...${NC}"
        helm upgrade --install claude-pipeline ./deploy/helm/claude-pipeline \
            --namespace $NAMESPACE \
            --create-namespace \
            --set image.tag=$VERSION

        echo ""
        echo -e "${GREEN}Deployment successful!${NC}"
        helm status claude-pipeline -n $NAMESPACE
        ;;

    *)
        echo -e "${RED}Unknown deployment type: $DEPLOY_TYPE${NC}"
        echo ""
        echo "Usage: $0 [docker|docker-prod|kubernetes|helm]"
        exit 1
        ;;
esac