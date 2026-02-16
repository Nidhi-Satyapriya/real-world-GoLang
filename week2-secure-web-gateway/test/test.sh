#!/bin/bash

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==================================${NC}"
echo -e "${BLUE}  Secure Web Gateway Test Suite${NC}"
echo -e "${BLUE}==================================${NC}"
echo ""

# Function to check if service is running
check_service() {
    local url=$1
    local name=$2
    
    echo -n "Checking $name... "
    if curl -s "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Running${NC}"
        return 0
    else
        echo -e "${RED}✗ Not running${NC}"
        return 1
    fi
}

# Function to test HTTP status
test_status() {
    local url=$1
    local expected=$2
    local description=$3
    
    echo -n "  Testing $description... "
    status=$(curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    
    if [ "$status" = "$expected" ]; then
        echo -e "${GREEN}✓ OK${NC} (HTTP $status)"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected $expected, got $status)"
        return 1
    fi
}

# Check if services are running
echo -e "${YELLOW}1. Service Health Checks${NC}"
echo ""

policy_running=false
proxy_running=false

if check_service "http://localhost:8000/health" "Policy Engine"; then
    policy_running=true
fi

if check_service "http://localhost:8080" "Proxy Server (expected connection error)"; then
    proxy_running=true
fi

echo ""

if ! $policy_running || ! $proxy_running; then
    echo -e "${RED}Error: Services are not running!${NC}"
    echo ""
    echo "Please start the services:"
    echo "  Terminal 1: make policy"
    echo "  Terminal 2: make proxy"
    echo ""
    exit 1
fi

# Test Policy Engine API
echo -e "${YELLOW}2. Policy Engine API Tests${NC}"
echo ""

echo -n "  GET /policy... "
if curl -s http://localhost:8000/policy | grep -q "blocked"; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
fi

echo -n "  GET /policy/domains... "
if curl -s http://localhost:8000/policy/domains | grep -q "total"; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
fi

echo ""

# Test blocked domains
echo -e "${YELLOW}3. Blocked Domains (should return 403)${NC}"
echo ""

test_status "http://facebook.com" "403" "facebook.com"
test_status "http://tiktok.com" "403" "tiktok.com"
test_status "http://youtube.com" "403" "youtube.com"
test_status "http://twitter.com" "403" "twitter.com"

echo ""

# Test allowed domains
echo -e "${YELLOW}4. Allowed Domains (should return 200/301/302)${NC}"
echo ""

echo -n "  Testing google.com... "
status=$(curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" "http://google.com" 2>/dev/null)
if [[ "$status" =~ ^(200|301|302)$ ]]; then
    echo -e "${GREEN}✓ OK${NC} (HTTP $status)"
else
    echo -e "${RED}✗ FAILED${NC} (HTTP $status)"
fi

echo -n "  Testing github.com... "
status=$(curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" "http://github.com" 2>/dev/null)
if [[ "$status" =~ ^(200|301|302)$ ]]; then
    echo -e "${GREEN}✓ OK${NC} (HTTP $status)"
else
    echo -e "${RED}✗ FAILED${NC} (HTTP $status)"
fi

echo ""

# Test dynamic policy updates
echo -e "${YELLOW}5. Dynamic Policy Management${NC}"
echo ""

echo -n "  Adding 'example.com' to blocklist... "
if curl -s -X POST "http://localhost:8000/policy/add?domain=example.com" | grep -q "added"; then
    echo -e "${GREEN}✓ OK${NC}"
    
    sleep 1
    echo -n "  Verifying in policy... "
    if curl -s http://localhost:8000/policy | grep -q "example.com"; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
    fi
    
    echo -n "  Removing 'example.com' from blocklist... "
    if curl -s -X DELETE "http://localhost:8000/policy/remove?domain=example.com" | grep -q "removed"; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
    fi
else
    echo -e "${RED}✗ FAILED${NC}"
fi

echo ""

# Test subdomain matching
echo -e "${YELLOW}6. Subdomain Matching${NC}"
echo ""

test_status "http://www.facebook.com" "403" "www.facebook.com (subdomain)"
test_status "http://m.facebook.com" "403" "m.facebook.com (subdomain)"

echo ""

# Summary
echo -e "${BLUE}==================================${NC}"
echo -e "${GREEN}Test suite completed!${NC}"
echo -e "${BLUE}==================================${NC}"
echo ""
echo "For more detailed testing, see:"
echo "  - docs/README.md"
echo "  - docs/QUICKSTART.md"
echo ""
