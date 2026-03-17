# Siriusec MCP Server

Siriusec MCP 是一个为 AI 系统提供 MCP（Model Context Protocol）协议接入能力的 Go 语言服务器。

## 项目概述

Siriusec MCP 实现了 MCP 标准协议的统一网关，支持多种传输模式，并深度集成阿里云 OpenAPI 生态（SysOM 监控服务）。

### 核心特性

- **MCP 协议网关**: 统一接入点 `/mcp/unified`
- **多传输模式支持**: stdio、SSE、Streamable-HTTP
- **阿里云生态集成**: SysOM 监控、OpenAPI、DashScope
- **系统诊断工具**: 内存、IO、网络、调度等多维度诊断
- **崩溃诊断代理**: VMCORE 分析与诊断任务管理

## 功能模块

```
Siriusec_MCP 网关
├── MCP 协议网关
│   ├── 协议解析
│   ├── 统一接入 /mcp/unified
│   └── 流式响应 (SSE / Streamable-HTTP)
├── 云服务集成
│   ├── 阿里云 SysOM 监控
│   ├── 通用 OpenAPI
│   └── 通义千问 DashScope
└── 系统诊断工具
    ├── AM 工具 (实例/集群/Pod 管理)
    ├── CrashAgent (崩溃诊断)
    ├── IO 诊断 (iofsstat/iodiagnose)
    ├── 内存诊断 (memgraph/javamem/oomcheck)
    ├── 网络诊断 (packetdrop/netjitter)
    └── 调度诊断 (delay/loadtask)
```

## 快速开始

### 环境要求

- Go 1.25+
- Docker (可选)
- 阿里云账号及 AccessKey

### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd siriusec-mcp

# 构建
./build.sh local
# 或
go build -o bin/siriusec-mcp cmd/server/main.go
```

### 配置

创建 `.env` 文件：

```env
# OpenAPI 配置
OPENAPI_TYPE=access_key
OPENAPI_ACCESS_KEY_ID=your-access-key-id
OPENAPI_ACCESS_KEY_SECRET=your-access-key-secret
REGION_ID=cn-hangzhou

# 可选: LLM 配置
DASHSCOPE_API_KEY=your-dashscope-api-key

# 日志级别
LOG_LEVEL=INFO
```

验证配置：
```bash
./bin/siriusec-mcp config validate
```

### 运行

```bash
# stdio 模式 (默认)
./bin/siriusec-mcp run

# SSE 模式
./bin/siriusec-mcp run --sse --host 0.0.0.0 --port 7140

# Streamable-HTTP 模式
./bin/siriusec-mcp run --streamable-http --host 0.0.0.0 --port 7140
```

## 工具列表

| 分类 | 工具名称 | 描述 |
|------|---------|------|
| **AM** | `list_all_instances` | 列出所有实例 |
| | `list_instances` | 列出实例列表 |
| | `list_clusters` | 列出集群列表 |
| | `list_pods_of_instance` | 列出实例下的 Pod |
| **内存诊断** | `memgraph` | 内存全景分析 |
| | `javamem` | Java 内存诊断 |
| | `oomcheck` | OOM 诊断 |
| **IO 诊断** | `iofsstat` | IO 流量分析 |
| | `iodiagnose` | IO 一键诊断 |
| **网络诊断** | `packetdrop` | 网络丢包诊断 |
| | `netjitter` | 网络抖动诊断 |
| **调度诊断** | `delay` | 调度抖动诊断 |
| | `loadtask` | 系统负载诊断 |
| **其他诊断** | `vmcore_analysis` | 宕机诊断 |
| | `disk_analysis` | 磁盘分析 |
| **CrashAgent** | `list_vmcores` | 查询历史诊断任务 |
| | `get_vmcore_detail` | 获取诊断任务详情 |
| | `analyze_vmcore` | 创建 VMCORE 诊断任务 |
| | `delete_vmcore` | 创建 dmesg 诊断任务 |
| **初始化** | `check_sysom_initialed` | 检查 SysOM 开通状态 |
| | `initial_sysom` | 开通 SysOM 服务 |

## CLI 命令

Siriusec MCP 提供强大的命令行工具：

```bash
# 查看版本
./bin/siriusec-mcp version

# 列出所有工具
./bin/siriusec-mcp tools list
./bin/siriusec-mcp tools list --format json
./bin/siriusec-mcp tools list --format table

# 验证配置
./bin/siriusec-mcp config validate

# 自测工具
./bin/siriusec-mcp selftest
./bin/siriusec-mcp selftest --filter am
./bin/siriusec-mcp selftest -v

# 测试 MCP 服务器连接
./bin/siriusec-mcp mcptest
./bin/siriusec-mcp mcptest --url http://remote:7140

# 查看帮助
./bin/siriusec-mcp help
```

## Docker 部署

```bash
# 构建镜像
docker build -t siriusec-mcp:latest .

# 运行 (stdio 模式)
docker run -i --rm \
  -e OPENAPI_ACCESS_KEY_ID=$ACCESS_KEY_ID \
  -e OPENAPI_ACCESS_KEY_SECRET=$ACCESS_KEY_SECRET \
  siriusec-mcp:latest

# 运行 (SSE 模式)
docker run -d \
  -p 7140:7140 \
  -e OPENAPI_ACCESS_KEY_ID=$ACCESS_KEY_ID \
  -e OPENAPI_ACCESS_KEY_SECRET=$ACCESS_KEY_SECRET \
  siriusec-mcp:latest \
  run --sse --host 0.0.0.0 --port 7140
```

## 项目结构

```
.
├── cmd/server/          # 主程序入口
├── internal/
│   ├── cli/             # CLI 命令实现
│   ├── client/          # OpenAPI 客户端
│   ├── config/          # 配置管理
│   ├── logger/          # 日志模块 (zap)
│   ├── mcp/             # MCP 服务器核心
│   ├── registry/        # API 路由注册表
│   ├── tools/           # 工具实现
│   │   ├── am/          # 应用管理工具
│   │   ├── crashagent/  # 崩溃诊断代理
│   │   ├── diagnosis/   # 诊断基类
│   │   ├── initial/     # 初始化服务
│   │   ├── iodiag/      # IO 诊断
│   │   ├── memdiag/     # 内存诊断
│   │   ├── netdiag/     # 网络诊断
│   │   ├── otherdiag/   # 其他诊断
│   │   └── scheddiag/   # 调度诊断
│   └── version/         # 版本信息
├── pkg/models/          # 数据模型
├── test/                # 测试
│   ├── integration/     # 集成测试
│   └── unit/            # 单元测试
├── build.sh             # 多平台构建脚本
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行单元测试
go test ./test/unit/...

# 运行集成测试
go test ./test/integration/...

# 带覆盖率
go test -cover ./...
```

## 技术栈

- **语言**: Go 1.25
- **MCP 协议**: [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)
- **阿里云 SDK**: [alibabacloud-go/sysom-20231230](https://github.com/alibabacloud-go/sysom-20231230)
- **日志**: [uber-go/zap](https://github.com/uber-go/zap) (结构化 JSON 日志)
- **配置**: [joho/godotenv](https://github.com/joho/godotenv)
- **构建**: 多平台构建脚本 (build.sh)

## 许可证

[MIT License](LICENSE)
