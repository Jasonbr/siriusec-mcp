# OpenClaw 快速接入指南 - 无需 Skill

## 🎯 推荐方案：直接调用 API（无需编写 Skill）

### 为什么不需要 Skill？

Siriusec MCP 已经内置了完整的 LLM 客户端，支持：
- ✅ OpenAI兼容协议
- ✅ 自定义端点
- ✅ 多模型切换
- ✅ 自动重试和超时处理

**您只需要配置环境变量即可！**

---

## 🚀 三步快速接入

### 步骤 1：复制配置模板

```bash
cd /Users/xiaoxi/Downloads/workspace/siriusec_mcp
cp .env.openclaw .env
```

### 步骤 2：编辑配置文件

打开 `.env` 文件，修改以下内容：

```bash
# LLM 提供商（OpenClaw 使用 OpenAI兼容协议）
LLM_PROVIDER=openai

# OpenClaw API Key（如果没有可留空或填任意值）
LLM_API_KEY=not-needed

# OpenClaw API 端点（根据您的部署修改）
# 本地部署示例：
LLM_BASE_URL=http://localhost:8000/v1/chat/completions

# 远程部署示例：
# LLM_BASE_URL=http://your-server:8000/v1/chat/completions

# 模型名称（根据 OpenClaw 配置的模型填写）
LLM_MODEL=qwen-7b  # 或其他您配置的模型

# 温度参数 (0-1, 越高越有创造性)
LLM_TEMPERATURE=0.7

# 最大 Token 数
LLM_MAX_TOKENS=4000

# 请求超时时间（秒）
LLM_TIMEOUT=120
```

### 步骤 3：启动服务

```bash
# 方式 1：一键启动
./start-webui.sh

# 方式 2：分步启动
# 终端 1 - 启动 MCP 服务器
./bin/siriusec-mcp run --streamable-http --port 7140

# 终端 2 - 启动 Web UI
./bin/siriusec-webui -port 8080
```

---

## ✅ 验证连接

### 测试 1：检查配置

```bash
./bin/siriusec-mcp config validate
```

### 测试 2：测试 LLM 连接

```bash
# 创建测试脚本
cat > test_openclaw.sh << 'TEST'
#!/bin/bash

echo "测试 OpenClaw 连接..."

# 从 .env 读取配置
source .env

# 发送测试请求
curl -X POST "$LLM_BASE_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $LLM_API_KEY" \
  -d '{
    "model": "'"$LLM_MODEL"'",
    "messages": [
      {"role": "user", "content": "你好，请简单介绍一下自己"}
    ],
    "max_tokens": 100
  }' | jq .

echo ""
echo "如果看到响应，说明连接成功！"
TEST

chmod +x test_openclaw.sh
./test_openclaw.sh
```

### 测试 3：通过 Web UI 测试

1. 访问 http://localhost:8080
2. 在对话框输入："你好，帮我诊断一下服务器"
3. 如果 AI 回复，说明集成成功！

---

## 🔧 OpenClaw 部署参考

### 本地部署（Docker）

```bash
# 部署 OpenClaw（以 Qwen 模型为例）
docker run -d \
  --gpus all \
  -p 8000:8000 \
  -v /path/to/models:/models \
  -e MODEL_PATH=/models/qwen-7b \
  openclaw/openclaw:latest

# 等待模型加载完成（约 1-2 分钟）
docker logs -f <container_id>
```

### 配置 Siriusec MCP

```bash
# 编辑 .env
LLM_BASE_URL=http://localhost:8000/v1/chat/completions
LLM_MODEL=qwen-7b
```

---

## 💡 工作原理

```
用户提问
   ↓
Web UI 接收
   ↓
MCP 服务器处理
   ↓
LLM 客户端封装请求
   ↓
发送到 OpenClaw API ← 你在这里配置端点
   ↓
OpenClaw 返回响应
   ↓
MCP 解析并返回结果
   ↓
Web UI 显示答案
```

**整个过程不需要编写任何 Skill！**

---

## 🛠️ 常见问题

### Q1: OpenClaw 需要 API Key 吗？
A: 通常不需要。可以设置 `LLM_API_KEY=not-needed` 或任意值。

### Q2: 如何知道 OpenClaw 的端点？
A: 默认是 `http://localhost:8000/v1/chat/completions`

### Q3: 支持哪些模型？
A: 任何 OpenClaw 支持的模型都可以：
- Qwen 系列
- Llama 系列
- ChatGLM 系列
- Baichuan 系列
- 等等...

### Q4: 响应很慢怎么办？
A: 调整超时时间：
```bash
LLM_TIMEOUT=300  # 增加到 5 分钟
```

### Q5: 可以同时使用多个模型吗？
A: 可以！创建多个 `.env` 文件切换：
```bash
.env.qwen    # Qwen 模型
.env.llama   # Llama 模型
cp .env.qwen .env  # 切换
./start-webui.sh   # 重启
```

---

## 📊 性能优化建议

### 1. 调整并发数
在 `cmd/server/main.go` 中：
```go
maxConcurrency := 20  // 增加并发数
```

### 2. 使用 SSE模式
```bash
./bin/siriusec-mcp run --sse --port 7140
```

### 3. 配置反向代理（生产环境）
使用 Nginx 提供 HTTPS 和负载均衡：
```nginx
location /mcp {
    proxy_pass http://localhost:7140;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

---

## 🎉 总结

✅ **不需要编写 Skill**
✅ **只需配置环境变量**
✅ **3 步即可完成集成**
✅ **支持任意 OpenAI兼容模型**
✅ **开箱即用**

立即体验：
```bash
./start-webui.sh
# 访问 http://localhost:8080
```
