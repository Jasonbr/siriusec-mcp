#!/bin/bash
# 测试阿里百炼模型

echo "=========================================="
echo "测试阿里百炼 kimi-k2.5 模型"
echo "=========================================="
echo ""

# 检查配置
echo "1. 检查配置..."
./bin/siriusec-mcp config validate | grep -A5 "LLM"
echo ""

# 启动服务器（后台）
echo "2. 启动 MCP 服务器..."
./bin/siriusec-mcp run --streamable-http --port 7150 &
SERVER_PID=$!
sleep 3

# 测试健康检查
echo "3. 测试健康检查..."
curl -s http://localhost:7150/health | jq .
echo ""

# 测试 LLM 工具
echo "4. 测试 LLM 智能诊断工具..."
curl -s -X POST http://localhost:7150/mcp/unified \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "smart_diagnose",
      "arguments": {
        "symptom": "服务器 CPU 使用率突然飙升到 100%",
        "context": "阿里云 ECS 实例，运行 Java 应用，最近没有发布新版本"
      }
    }
  }' | jq .

echo ""
echo "5. 停止服务器..."
kill $SERVER_PID 2>/dev/null

echo ""
echo "测试完成！"
