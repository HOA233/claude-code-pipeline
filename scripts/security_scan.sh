#!/bin/bash
# Security Scan Script for Claude Pipeline

set -e

echo "============================================"
echo "Security Scan for Claude Pipeline"
echo "============================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ISSUES_FOUND=0

# Check for sensitive data exposure
check_sensitive_data() {
    echo ""
    echo "=== Checking for Sensitive Data Exposure ==="

    # Check for API keys in code
    if grep -r "api_key\s*=\s*['\"][^'\"]*['\"]" --include="*.go" --include="*.js" --include="*.jsx" . 2>/dev/null | grep -v "test" | grep -v "example"; then
        echo -e "${RED}✗ Hardcoded API keys found${NC}"
        ((ISSUES_FOUND++))
    else
        echo -e "${GREEN}✓ No hardcoded API keys found${NC}"
    fi

    # Check for passwords in code
    if grep -r "password\s*=\s*['\"][^'\"]*['\"]" --include="*.go" --include="*.js" --include="*.jsx" . 2>/dev/null | grep -v "test" | grep -v "example"; then
        echo -e "${RED}✗ Hardcoded passwords found${NC}"
        ((ISSUES_FOUND++))
    else
        echo -e "${GREEN}✓ No hardcoded passwords found${NC}"
    fi

    # Check for secrets in YAML files
    if grep -r "secret\s*:\s*['\"][^'\"]*['\"]" --include="*.yaml" --include="*.yml" . 2>/dev/null | grep -v "example"; then
        echo -e "${YELLOW}⚠ Potential secrets in YAML files${NC}"
    fi
}

# Check for SQL injection vulnerabilities
check_sql_injection() {
    echo ""
    echo "=== Checking for SQL Injection Risks ==="

    if grep -r "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE\|fmt.Sprintf.*DELETE" --include="*.go" . 2>/dev/null; then
        echo -e "${RED}✗ Potential SQL injection vulnerabilities found${NC}"
        ((ISSUES_FOUND++))
    else
        echo -e "${GREEN}✓ No obvious SQL injection patterns found${NC}"
    fi
}

# Check for command injection vulnerabilities
check_command_injection() {
    echo ""
    echo "=== Checking for Command Injection Risks ==="

    if grep -r "exec.Command.*fmt.Sprintf\|exec.Command.*+" --include="*.go" . 2>/dev/null | head -5; then
        echo -e "${YELLOW}⚠ Review command execution patterns for injection risks${NC}"
    else
        echo -e "${GREEN}✓ No obvious command injection patterns${NC}"
    fi
}

# Check for insecure dependencies
check_dependencies() {
    echo ""
    echo "=== Checking Dependencies ==="

    # Check Go dependencies
    if command -v govulncheck &> /dev/null; then
        echo "Running govulncheck..."
        govulncheck ./... 2>/dev/null || echo -e "${YELLOW}⚠ Some vulnerabilities may exist in dependencies${NC}"
    else
        echo -e "${YELLOW}⚠ govulncheck not installed, skipping Go dependency check${NC}"
    fi

    # Check npm dependencies
    if [ -f "frontend/package.json" ]; then
        echo "Checking npm dependencies..."
        cd frontend && npm audit --audit-level=moderate 2>/dev/null || true
        cd ..
    fi
}

# Check for proper authentication
check_authentication() {
    echo ""
    echo "=== Checking Authentication Implementation ==="

    # Check for authentication middleware
    if grep -r "func.*Auth\|func.*auth" --include="*.go" ./internal/api 2>/dev/null | head -3; then
        echo -e "${GREEN}✓ Authentication functions found${NC}"
    else
        echo -e "${YELLOW}⚠ No obvious authentication implementation found${NC}"
    fi

    # Check for rate limiting
    if grep -r "RateLimiter\|rate_limit" --include="*.go" . 2>/dev/null | head -1; then
        echo -e "${GREEN}✓ Rate limiting implementation found${NC}"
    else
        echo -e "${YELLOW}⚠ No rate limiting found${NC}"
    fi
}

# Check for proper CORS configuration
check_cors() {
    echo ""
    echo "=== Checking CORS Configuration ==="

    if grep -r "AllowAllOrigins\|AllowOriginFunc" --include="*.go" . 2>/dev/null | head -3; then
        echo -e "${YELLOW}⚠ Review CORS configuration for production${NC}"
    else
        echo -e "${GREEN}✓ No overly permissive CORS found${NC}"
    fi
}

# Check for secure headers
check_security_headers() {
    echo ""
    echo "=== Checking Security Headers ==="

    if grep -r "X-Content-Type-Options\|X-Frame-Options\|X-XSS-Protection" --include="*.go" . 2>/dev/null | head -1; then
        echo -e "${GREEN}✓ Security headers implementation found${NC}"
    else
        echo -e "${YELLOW}⚠ Consider adding security headers middleware${NC}"
    fi
}

# Check Kubernetes security
check_k8s_security() {
    echo ""
    echo "=== Checking Kubernetes Security ==="

    if [ -d "deploy/kubernetes" ]; then
        # Check for privileged containers
        if grep -r "privileged:\s*true" deploy/kubernetes 2>/dev/null; then
            echo -e "${RED}✗ Privileged containers found${NC}"
            ((ISSUES_FOUND++))
        else
            echo -e "${GREEN}✓ No privileged containers${NC}"
        fi

        # Check for root user
        if grep -r "runAsUser:\s*0" deploy/kubernetes 2>/dev/null; then
            echo -e "${YELLOW}⚠ Containers running as root${NC}"
        fi

        # Check for resource limits
        if grep -r "resources:" deploy/kubernetes 2>/dev/null | head -1; then
            echo -e "${GREEN}✓ Resource limits defined${NC}"
        else
            echo -e "${YELLOW}⚠ Consider adding resource limits${NC}"
        fi
    fi
}

# Run all checks
check_sensitive_data
check_sql_injection
check_command_injection
check_dependencies
check_authentication
check_cors
check_security_headers
check_k8s_security

echo ""
echo "============================================"
echo "Security Scan Complete"
echo "============================================"

if [ $ISSUES_FOUND -gt 0 ]; then
    echo -e "${RED}Issues found: $ISSUES_FOUND${NC}"
    exit 1
else
    echo -e "${GREEN}No critical security issues found${NC}"
    exit 0
fi