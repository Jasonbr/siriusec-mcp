#!/bin/bash

# Siriusec MCP 测试脚本

set -e

echo "🧪 Testing Siriusec MCP..."
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数器
TESTS_PASSED=0
TESTS_FAILED=0

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local expected_code=$3
    
    echo -n "Testing $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✅ PASSED${NC} (HTTP $response)"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED${NC} (Expected: $expected_code, Got: $response)"
        ((TESTS_FAILED++))
    fi
}

test_json_api() {
    local name=$1
    local url=$2
    local method=$3
    local data=$4
    
    echo -n "Testing $name... "
    
    response=$(curl -s -X "$method" -H "Content-Type: application/json" -d "$data" "$url" 2>/dev/null)
    
    if echo "$response" | grep -q "result\|error\|content" > /dev/null; then
        echo -e "${GREEN}✅ PASSED${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED${NC} (Invalid JSON response)"
        ((TESTS_FAILED++))
    fi
}

# 等待服务启动
echo "⏳ Waiting for services to be ready..."
sleep 2

echo ""
echo "=========================================="
echo "📋 Running Tests"
echo "=========================================="
echo ""

# Test 1: MCP Server Health Check
test_endpoint "MCP Server Health" "http://localhost:7140/health" "200"

# Test 2: Web UI (if running)
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    test_endpoint "Web UI" "http://localhost:3000" "200"
else
    echo -e "${YELLOW}⚠️  SKIPPED${NC} Web UI (not running)"
fi

# Test 3: MCP Tools List
test_json_api "MCP Tools List" \
    "http://localhost:7140/mcp/unified" \
    "POST" \
    '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Test 4: System Metrics (CPU Plugin)
echo ""
echo "Testing monitoring plugins..."
test_endpoint "CPU Monitoring" "http://localhost:7140/plugins/cpu" "200"

# Test 5: Memory Plugin
test_endpoint "Memory Monitoring" "http://localhost:7140/plugins/mem" "200"

# Test 6: Disk Plugin
test_endpoint "Disk Monitoring" "http://localhost:7140/plugins/disk" "200"

# Summary
echo ""
echo "=========================================="
echo "📊 Test Summary"
echo "=========================================="
echo ""
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
fi
