# OpenClaw 完整接入指南

## 📋 您的配置信息

### 本地网关
- **端口**: 18789
- **模式**: local
- **绑定**: lan
- **认证**: token (xiaoxi)

### 云端服务
- **Web UI**: https://siriusec.com/claw/chat
- **WebSocket**: wss://siriusec.com/claw/api/ws
- **Token**: xiaoxi
- **默认会话**: agent:main:main

---

## 🚀 快速开始（推荐：使用本地网关）

### 步骤 1：应用本地配置

```bash
cd /Users/xiaoxi/Downloads/workspace/siriusec_mcp

# 使用本地网关配置
cp .env.openclaw-local .env
```

### 步骤 2：确认模型名称

编辑 `.env` 文件，修改 `LLM_MODEL` 为您实际使用的模型：

```bash
vim .env
# 修改第 30 行：LLM_MODEL=your-model-name
```

常见模型：
- `qwen-7b`
- `llama-2-7b`
- `chatglm3-6b`
- `baichuan2-7b`

### 步骤 3：测试连接

```bash
chmod +x test-openclaw.sh
./test-openclaw.sh
```

### 步骤 4：启动服务

```bash
chmod +x start-webui.sh
./start-webui.sh
```

### 步骤 5：访问 Web UI

打开浏览器访问：http://localhost:8080

---

## ☁️ 备选方案：使用云端服务

如果您想使用云端 OpenClaw 服务：

```bash
# 使用云端配置
cp .env.openclaw-cloud .env

# 注意：需要确认云端 API 端点格式
# 可能需要调整 LLM_BASE_URL
```

---

## 🔧 配置说明

### 本地网关 vs 云端服务

| 特性 | 本地网关 | 云端服务 |
|------|---------|---------|
| **延迟** | ⭐⭐⭐⭐⭐ 极低 | ⭐⭐⭐ 依赖网络 |
| **数据隐私** | ⭐⭐⭐⭐⭐ 完全本地 | ⭐⭐⭐ 数据传输 |
| **可访问性** | ⭐⭐⭐ 仅本地 | ⭐⭐⭐⭐⭐  anywhere |
| **维护成本** | ⭐⭐⭐⭐ 自己维护 | ⭐⭐⭐⭐⭐ 无需维护 |

### 推荐场景

**使用本地网关：**
- ✅ 生产环境
- ✅ 数据敏感场景
- ✅ 需要低延迟
- ✅ 大规模调用

**使用云端服务：**
- ✅ 开发测试
- ✅ 多地点协作
- ✅ 临时使用
- ✅ 不想维护基础设施

---

## 🧪 测试连接

### 测试本地网关

```bash
# 发送测试请求
curl -X POST "http://localhost:18789/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer xiaoxi" \
  -d '{
    "model": "qwen-7b",
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "max_tokens": 50
  }' | jq .
```

### 预期响应

```json
{
  "choices": [
    {
      "message": {
        "content": "你好！我是 AI 助手..."
      }
    }
  ]
}
```

---

## ⚠️ 常见问题

### Q1: 无法连接到本地网关？

**检查 OpenClaw 是否运行：**

```bash
# Docker 部署
docker ps | grep openclaw

# 或直接访问
curl http://localhost:18789/
```

**如果未运行，启动 OpenClaw：**

```bash
# 根据您的部署方式启动
docker-compose up -d openclaw
# 或
openclaw start --config config.yaml
```

### Q2: 模型名称错误？

**查看 OpenClaw 支持的模型：**

```bash
# 方法 1：查看配置文件
cat /path/to/openclaw/config.yaml

# 方法 2：调用 models API
curl http://localhost:18789/v1/models \
  -H "Authorization: Bearer xiaoxi" | jq .
```

### Q3: 认证失败？

确认 Token 正确：

```bash
# 在 .env 文件中
LLM_API_KEY=xiaoxi  # 确保与 auth.token 一致
```

### Q4: WebSocket 如何使用？

当前 Siriusec MCP 使用 HTTP REST API，不支持 WebSocket。

如果需要 WebSocket 支持，可以：
1. 使用官方的 Web UI：https://siriusec.com/claw/chat
2. 自行实现 WebSocket 客户端

---

## 💡 最佳实践

### 1. 开发环境配置

```bash
# 使用本地网关快速测试
cp .env.openclaw-local .env

# 设置较短的超时时间
LLM_TIMEOUT=60
```

### 2. 生产环境配置

```bash
# 增加并发和超时
LLM_MAX_TOKENS=8000
LLM_TIMEOUT=300

# 添加重试逻辑（在代码中）
```

### 3. 性能优化

```bash
# 调整温度参数控制创造性
LLM_TEMPERATURE=0.3  # 更确定性
LLM_TEMPERATURE=0.7  # 更有创造性

# 减少 Token 数提高速度
LLM_MAX_TOKENS=2000
```

---

## 📊 监控和日志

### 查看 MCP服务器日志

```bash
# 如果以后台服务运行
journalctl -u siriusec-mcp -f

# 或直接查看输出
./bin/siriusec-mcp run --streamable-http --port 7140
```

### 健康检查

```bash
# 检查 MCP服务器
curl http://localhost:7140/health | jq .

# 检查 OpenClaw
curl http://localhost:18789/health | jq .
```

---

## 🎉 总结

✅ **已创建配置文件：**
- `.env.openclaw-local` - 本地网关配置
- `.env.openclaw-cloud` - 云端服务配置

✅ **已创建测试脚本：**
- `test-openclaw.sh` - 自动测试连接

✅ **快速启动：**
```bash
cp .env.openclaw-local .env
./test-openclaw.sh
./start-webui.sh
```

立即体验：http://localhost:8080
