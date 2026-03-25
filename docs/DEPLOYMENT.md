# Siriusec MCP 部署文档

## 目录

- [环境要求](#环境要求)
- [本地部署](#本地部署)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [配置说明](#配置说明)
- [监控与健康检查](#监控与健康检查)
- [故障排查](#故障排查)

## 环境要求

### 基础要求

- **操作系统**: Linux, macOS, Windows (WSL2)
- **Go 版本**: 1.25 或更高
- **内存**: 最低 256MB，推荐 512MB
- **磁盘**: 最低 100MB 可用空间
- **网络**: 能够访问阿里云 OpenAPI 服务

### 阿里云账号要求

- 有效的阿里云账号
- 已开通 SysOM 服务
- 具备以下权限的 AccessKey:
  - `AliyunSysOMFullAccess`
  - `AliyunECSReadOnlyAccess` (推荐)

## 本地部署

### 1. 源码编译

```bash
# 克隆代码
git clone <repository-url>
cd siriusec-mcp

# 下载依赖
go mod download

# 编译（使用构建脚本）
./build.sh local

# 或手动编译
go build -o bin/siriusec-mcp cmd/server/main.go
```

### 2. 配置环境变量

创建 `.env` 文件：

```env
# 必需配置
OPENAPI_ACCESS_KEY_ID=your-access-key-id
OPENAPI_ACCESS_KEY_SECRET=your-access-key-secret
REGION_ID=cn-hangzhou

# 可选配置
LOG_LEVEL=INFO
DASHSCOPE_API_KEY=your-dashscope-api-key
```

验证配置：
```bash
./bin/siriusec-mcp config validate
```

### 3. 运行服务

#### stdio 模式（默认）

适用于本地 AI 客户端直接调用：

```bash
./bin/siriusec-mcp run
```

#### SSE 模式

适用于 HTTP 长连接场景：

```bash
./bin/siriusec-mcp run \
  --sse \
  --host 0.0.0.0 \
  --port 7140 \
  --path /mcp/unified
```

访问地址: `http://localhost:7140/mcp/unified`

#### Streamable-HTTP 模式

适用于无状态 HTTP 请求：

```bash
./bin/siriusec-mcp run \
  --streamable-http \
  --host 0.0.0.0 \
  --port 7140 \
  --path /mcp/unified
```

## Docker 部署

### 1. 构建镜像

```bash
docker build -t siriusec-mcp:latest .
```

### 2. 运行容器

#### stdio 模式

```bash
docker run -i --rm \
  -e OPENAPI_ACCESS_KEY_ID=$ACCESS_KEY_ID \
  -e OPENAPI_ACCESS_KEY_SECRET=$ACCESS_KEY_SECRET \
  -e REGION_ID=cn-hangzhou \
  siriusec-mcp:latest
```

#### SSE 模式

```bash
docker run -d \
  --name siriusec-mcp-sse \
  -p 7140:7140 \
  -e OPENAPI_ACCESS_KEY_ID=$ACCESS_KEY_ID \
  -e OPENAPI_ACCESS_KEY_SECRET=$ACCESS_KEY_SECRET \
  -e REGION_ID=cn-hangzhou \
  -e LOG_LEVEL=INFO \
  siriusec-mcp:latest \
  run --sse --host 0.0.0.0 --port 7140 --path /mcp/unified
```

#### Streamable-HTTP 模式

```bash
docker run -d \
  --name siriusec-mcp-http \
  -p 7140:7140 \
  -e OPENAPI_ACCESS_KEY_ID=$ACCESS_KEY_ID \
  -e OPENAPI_ACCESS_KEY_SECRET=$ACCESS_KEY_SECRET \
  -e REGION_ID=cn-hangzhou \
  siriusec-mcp:latest \
  run --streamable-http --host 0.0.0.0 --port 7140 --path /mcp/unified
```

### 3. 使用 Docker Compose

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，填入你的配置
vim .env

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### Docker Compose 配置示例

```yaml
version: '3.8'

services:
  siriusec-mcp:
    build:
      context: .
      dockerfile: Dockerfile
    image: siriusec-mcp:latest
    container_name: siriusec-mcp
    restart: unless-stopped
    ports:
      - "7140:7140"
    environment:
      - DEPLOY_MODE=alibabacloud_sdk
      - OPENAPI_TYPE=access_key
      - OPENAPI_ACCESS_KEY_ID=${ACCESS_KEY_ID}
      - OPENAPI_ACCESS_KEY_SECRET=${ACCESS_KEY_SECRET}
      - REGION_ID=${REGION_ID:-cn-hangzhou}
      - LOG_LEVEL=${LOG_LEVEL:-INFO}
    command:
      - "run"
      - "--streamable-http"
      - "--host"
      - "0.0.0.0"
      - "--port"
      - "7140"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:7140/health", "||", "exit", "1"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

## Kubernetes 部署

### 1. 创建 Namespace

```bash
kubectl apply -f k8s/namespace.yaml
```

### 2. 创建 ConfigMap 和 Secret

```bash
# 配置 ConfigMap
kubectl apply -f k8s/configmap.yaml

# 配置 Secret（需要提前设置好值）
kubectl create secret generic siriusec-mcp-secret \
  --namespace=siriusec-mcp \
  --from-literal=access-key-id=your-access-key-id \
  --from-literal=access-key-secret=your-access-key-secret
```

### 3. 部署应用

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### 4. 配置 Ingress（可选）

```bash
kubectl apply -f k8s/ingress.yaml
```

### 5. 配置 HPA（可选）

```bash
kubectl apply -f k8s/hpa.yaml
```

### 完整部署

```bash
# 使用 kustomization 一键部署
kubectl apply -k k8s/

# 或者逐个部署
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/serviceaccount.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/ingress.yaml
```

## Web UI 部署

### 本地测试

```bash
# 进入 Web UI 目录
cd web-ui

# 启动本地 HTTP 服务器
python3 -m http.server 8080

# 或使用 Node.js
npx serve .
```

访问地址: `http://localhost:8080`

### 生产部署

Web UI 是静态文件，可以部署到任何静态文件服务器：

#### Nginx 配置示例

```nginx
server {
    listen 80;
    server_name siriusec-mcp.example.com;
    root /var/www/siriusec-mcp/web-ui;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # 反向代理 MCP API
    location /api/mcp/ {
        proxy_pass http://localhost:7140/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

#### 使用 Docker 部署 Web UI

```dockerfile
FROM nginx:alpine
COPY web-ui /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### Web UI 配置

编辑 `web-ui/index.html` 中的 MCP 服务器地址：

```javascript
// 开发环境（本地 MCP Server）
const MCP_SERVER_URL = 'http://localhost:7140/mcp/unified';

// 生产环境（通过 Nginx 反向代理）
const MCP_SERVER_URL = '/api/mcp/unified';
```

## 配置说明

### 环境变量

| 变量名 | 必填 | 默认值 | 描述 |
|--------|------|--------|------|
| `OPENAPI_ACCESS_KEY_ID` | 是 | - | 阿里云 AccessKey ID |
| `OPENAPI_ACCESS_KEY_SECRET` | 是 | - | 阿里云 AccessKey Secret |
| `OPENAPI_SECURITY_TOKEN` | 否 | - | STS 临时凭证 Token |
| `REGION_ID` | 否 | cn-hangzhou | 阿里云地域 |
| `DEPLOY_MODE` | 否 | alibabacloud_sdk | 部署模式 |
| `LOG_LEVEL` | 否 | INFO | 日志级别: DEBUG/INFO/WARN/ERROR |
| `DASHSCOPE_API_KEY` | 否 | - | 通义千问 API Key |

### 命令行参数

#### 服务器运行参数 (run 子命令)

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `--sse` | bool | false | SSE 模式 |
| `--streamable-http` | bool | false | Streamable-HTTP 模式 |
| `--host` | string | 127.0.0.1 | 监听地址 |
| `--port` | int | 7140 | 监听端口 |
| `--path` | string | /mcp/unified | 端点路径 |

#### CLI 命令

| 命令 | 描述 |
|------|------|
| `version` | 显示版本信息 |
| `help` | 显示帮助信息 |
| `tools list` | 列出所有工具 |
| `config validate` | 验证配置 |
| `selftest` | 运行自测 |
| `mcptest` | 测试 MCP 连接 |
| `run` | 启动服务器 |

### 配置文件优先级

1. 命令行参数（最高优先级）
2. 环境变量
3. `.env` 文件
4. 默认值（最低优先级）

## 监控与健康检查

### 健康检查端点

- **HTTP 模式**: `GET /health`
- **响应示例**:
  ```json
  {
    "status": "healthy",
    "version": "1.0.0",
    "timestamp": "2026-03-17T11:00:00Z"
  }
  ```

### Docker 健康检查

容器内置健康检查配置：

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7140/health || exit 1
```

### Kubernetes 探针

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 7140
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /health
    port: 7140
  initialDelaySeconds: 5
  periodSeconds: 10
```

### 日志监控

日志输出到 stderr，使用结构化 JSON 格式：

```json
{"level":"info","ts":"2026-03-17T11:00:00.000+0800","caller":"mcp/server.go:50","msg":"message"}
```

收集日志：

```bash
# Docker
docker logs siriusec-mcp

# Kubernetes
kubectl logs -f deployment/siriusec-mcp -n siriusec-mcp
```

## 故障排查

### 常见问题

#### 1. 启动失败：配置未加载

**症状**: `config not loaded` 错误

**解决**:
```bash
# 检查环境变量是否设置
echo $OPENAPI_ACCESS_KEY_ID
echo $OPENAPI_ACCESS_KEY_SECRET

# 检查 .env 文件是否存在且格式正确
ls -la .env
cat .env
```

#### 2. API 调用失败：权限错误

**症状**: `权限不足` 或 `unauthorized`

**解决**:
1. 检查 AccessKey 是否有效
2. 确认账号已开通 SysOM 服务
3. 为 RAM 用户添加 `AliyunSysOMFullAccess` 权限

```bash
# 检查 SysOM 开通状态
curl -X POST http://localhost:7140/mcp/unified \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "check_sysom_initialed",
      "arguments": {"uid": "your-user-id"}
    }
  }'
```

#### 3. 端口被占用

**症状**: `bind: address already in use`

**解决**:
```bash
# 查找占用进程
lsof -i :7140

# 终止进程或更换端口
./siriusec-mcp-server --port 7141
```

#### 4. Docker 容器无法访问

**症状**: 容器运行但无法连接

**解决**:
```bash
# 检查容器状态
docker ps -a
docker logs siriusec-mcp

# 检查端口映射
docker port siriusec-mcp

# 使用 mcptest 测试连接
./bin/siriusec-mcp mcptest --url http://localhost:7140

# 或使用 curl
curl http://localhost:7140/health
```

### 调试模式

启用 DEBUG 日志：

```bash
# 本地
LOG_LEVEL=DEBUG ./siriusec-mcp run

# Docker
docker run -e LOG_LEVEL=DEBUG siriusec-mcp:latest run --streamable-http
```

运行自测：

```bash
# 测试所有工具
./siriusec-mcp selftest

# 测试特定分类
./siriusec-mcp selftest --filter am

# 详细输出
./siriusec-mcp selftest -v

# 测试 OpenAPI 连接
./siriusec-mcp selftest --openapi
```

### 获取支持

- 查看日志: `docker logs` 或 `kubectl logs`
- 检查配置: `./siriusec-mcp config validate`
- 测试连接: `./siriusec-mcp mcptest` 或使用 curl
- 运行自测: `./siriusec-mcp selftest`
- 权限检查: 确认阿里云账号权限配置正确
