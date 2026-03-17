# Siriusec MCP API 文档

## 概述

Siriusec MCP Server 实现了 MCP (Model Context Protocol) 协议，提供统一的工具调用接口。支持三种传输模式：

- **stdio**: 标准输入输出（默认）
- **SSE**: Server-Sent Events
- **Streamable-HTTP**: 流式 HTTP

## 传输模式

### stdio 模式

适用于本地进程通信，AI 客户端通过 stdin/stdout 与服务器交互。

```bash
./siriusec-mcp run
```

### SSE 模式

适用于 HTTP 长连接场景。

```bash
./siriusec-mcp run --sse --host 0.0.0.0 --port 7140 --path /mcp/unified
```

**端点**:
- SSE 连接: `GET /mcp/unified`
- 消息发送: `POST /mcp/unified/message`

### Streamable-HTTP 模式

适用于无状态 HTTP 请求。

```bash
./siriusec-mcp run --streamable-http --host 0.0.0.0 --port 7140 --path /mcp/unified
```

**端点**:
- MCP 请求: `POST /mcp/unified`
- 健康检查: `GET /health`
- 版本信息: `GET /version`
- 就绪检查: `GET /ready`

## MCP 协议方法

### initialize

初始化连接，交换服务器和客户端能力。

**请求**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "client-name",
      "version": "1.0.0"
    }
  }
}
```

**响应**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "Siriusec MCP Server",
      "version": "1.0.0"
    }
  }
}
```

### tools/list

获取所有可用工具列表。

**请求**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

**响应**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "list_all_instances",
        "description": "列出所有实例，支持按地域、纳管类型、实例类型等条件筛选",
        "inputSchema": {
          "type": "object",
          "properties": {
            "uid": {"type": "string", "description": "用户ID"},
            "region": {"type": "string", "description": "地域，如cn-hangzhou"},
            "managedType": {"type": "string", "default": "all"},
            "instanceType": {"type": "string", "default": "ecs"}
          },
          "required": ["uid"]
        }
      }
    ]
  }
}
```

### tools/call

调用指定工具。

**请求**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "list_all_instances",
    "arguments": {
      "uid": "user-123",
      "region": "cn-hangzhou",
      "pageSize": "10"
    }
  }
}
```

**响应**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"code\":\"Success\",\"data\":{...},\"requestId\":\"xxx\"}"
      }
    ],
    "isError": false
  }
}
```

## 工具详细说明

### AM 工具 (应用管理)

#### list_all_instances

列出所有实例，支持多种筛选条件。

**参数**:
| 参数名 | 类型 | 必填 | 默认值 | 描述 |
|--------|------|------|--------|------|
| uid | string | 是 | - | 用户ID |
| region | string | 否 | - | 地域，如 cn-hangzhou |
| managedType | string | 否 | all | 纳管类型: managed/unmanaged/all/install/uninstall/upgrade |
| instanceType | string | 否 | ecs | 实例类型 |
| current | string | 否 | 1 | 页码 |
| pageSize | string | 否 | 10 | 每页数量 |
| maxResults | number | 否 | 100 | 最大结果数 |

#### list_instances

列出实例列表，支持按实例ID、状态、地域筛选。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| instance | string | 否 | 实例ID |
| status | string | 否 | 实例状态 |
| region | string | 否 | 地域 |
| clusterId | string | 否 | 集群ID |
| current | number | 否 | 页码 |
| pageSize | number | 否 | 每页数量 |

#### list_clusters

列出集群列表。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| name | string | 否 | 集群名称 |
| clusterId | string | 否 | 集群ID |
| clusterType | string | 否 | 集群类型 |
| clusterStatus | string | 否 | 集群状态 |

#### list_pods_of_instance

列出指定实例下的 Pod 列表。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| instance | string | 是 | 实例ID |
| clusterId | string | 否 | 集群ID |

### 内存诊断工具

#### memgraph

内存全景分析工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道: ecs/auto |
| instance | string | 否 | 实例ID |
| pod | string | 否 | Pod名称 |
| clusterType | string | 否 | 集群类型 |
| clusterId | string | 否 | 集群ID |
| namespace | string | 否 | Pod命名空间 |

#### javamem

Java 内存诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |
| pid | string | 否 | Java进程PID |
| duration | string | 否 | Profiling时长，默认0 |

#### oomcheck

OOM 诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 否 | 实例ID |
| pod | string | 否 | Pod名称 |
| time | string | 否 | 时间戳 |

### IO 诊断工具

#### iofsstat

IO 流量分析工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |
| timeout | string | 否 | 诊断时长，默认15 |
| disk | string | 否 | 磁盘名称 |

#### iodiagnose

IO 一键诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |
| timeout | string | 否 | 诊断时长，默认30秒 |

### 网络诊断工具

#### packetdrop

网络丢包诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |

#### netjitter

网络抖动诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |
| duration | string | 否 | 诊断时长，默认20秒 |
| threshold | string | 否 | 抖动阈值，默认10ms |

### 调度诊断工具

#### delay

调度抖动诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |
| duration | string | 否 | 诊断时长，默认20秒 |
| threshold | string | 否 | 抖动阈值，默认20ms |

#### loadtask

系统负载诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |

### 其他诊断工具

#### vmcore_analysis

宕机诊断工具，分析操作系统崩溃原因。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |

#### disk_analysis

磁盘分析诊断工具。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |
| region | string | 是 | 地域 |
| channel | string | 是 | 诊断通道 |
| instance | string | 是 | 实例ID |

### CrashAgent 工具

#### list_vmcores

查询历史创建的宕机诊断任务记录。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| days | number | 是 | 查询天数，1-30 |

#### get_vmcore_detail

查询诊断任务的执行状态和结果。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| task_id | string | 是 | 诊断任务ID |

#### analyze_vmcore

创建基于 VMCORE 文件的内核宕机诊断任务。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| vmcore_url | string | 是 | vmcore文件下载链接 |
| debuginfo_url | string | 否 | debuginfo文件下载链接 |
| debuginfo_common_url | string | 否 | debuginfo-common文件下载链接 |

#### delete_vmcore

创建基于 dmesg 日志的系统诊断任务。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| dmesg_url | string | 是 | dmesg日志文件下载链接 |

### 初始化服务工具

#### check_sysom_initialed

检查用户是否已开通 SysOM 服务。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |

#### initial_sysom

帮助用户开通 SysOM 服务。

**参数**:
| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| uid | string | 是 | 用户ID |

## CLI 命令参考

Siriusec MCP 提供强大的命令行工具，便于运维和调试。

### 基础命令

#### version

显示版本信息。

```bash
./siriusec-mcp version
```

**输出示例**:
```
 (commit: 3ec1bc8, built: 20260317130841, go: go1.25.0, os: darwin/amd64)
```

#### help

显示帮助信息。

```bash
./siriusec-mcp help
```

### 工具管理

#### tools list

列出所有可用工具。

```bash
# 简单格式（默认）
./siriusec-mcp tools list

# JSON 格式
./siriusec-mcp tools list --format json

# 表格格式
./siriusec-mcp tools list --format table
```

**输出示例**:
```
Total tools: 21

[sysom_am] (4 tools)
  ✓ list_all_instances - 列出所有实例，支持按地域、纳管类型、实例类型等条件筛选
  ✓ list_instances - 列出实例列表，支持按实例ID、状态、地域、集群等条件筛选
  ✓ list_clusters - 列出集群列表，支持按名称、类型、状态等条件筛选
  ✓ list_pods_of_instance - 列出指定实例下的Pod列表
...
```

### 配置管理

#### config validate

验证配置文件和环境变量。

```bash
./siriusec-mcp config validate
./siriusec-mcp config validate --dir /path/to/config
```

**输出示例**:
```
Validating configuration...

✓ Config directory: .
⚠ Environment file not found: ./.env

OpenAPI Configuration:
  ✗ ACCESS_KEY_ID is not set
  ✗ ACCESS_KEY_SECRET is not set
  ✓ Region: cn-hangzhou
  ✓ Type: access_key

Log Configuration:
  ✓ Level: INFO

Configuration validation completed.
```

### 测试命令

#### selftest

运行自测，检查所有工具是否正常。

```bash
# 运行所有测试
./siriusec-mcp selftest

# 只测试特定分类的工具
./bin/siriusec-mcp selftest --filter am

# 详细输出
./bin/siriusec-mcp selftest -v

# 测试 OpenAPI 连接
./bin/siriusec-mcp selftest --openapi
```

**输出示例**:
```
Running self-test...

Test Results: 21 passed, 0 failed, 21 total
```

#### mcptest

测试 MCP 服务器连接。

```bash
# 测试本地服务器
./siriusec-mcp mcptest

# 测试远程服务器
./siriusec-mcp mcptest --url http://remote:7140
```

**输出示例**:
```
Testing MCP server at http://127.0.0.1:7140

MCP Test Results:
-----------------
  ✓ Health check                   HTTP 200 (11.6ms)
  ✓ Version endpoint               HTTP 200 (1.6ms)
  ✓ Readiness check                HTTP 200 (477µs)
  ✓ MCP initialize                 HTTP 200 (1.2ms)

Results: 4 passed, 0 failed
```

### 服务器运行

#### run

启动 MCP 服务器。

```bash
# stdio 模式（默认）
./siriusec-mcp run

# SSE 模式
./siriusec-mcp run --sse --host 0.0.0.0 --port 7140

# Streamable-HTTP 模式
./siriusec-mcp run --streamable-http --host 0.0.0.0 --port 7140
```

**参数**:
| 参数 | 描述 | 默认值 |
|------|------|--------|
| `--sse` | 启用 SSE 模式 | false |
| `--streamable-http` | 启用 Streamable-HTTP 模式 | false |
| `--host` | 绑定主机 | 127.0.0.1 |
| `--port` | 绑定端口 | 7140 |
| `--path` | 端点路径 | /mcp/unified |

## 响应格式

### 成功响应

```json
{
  "code": "Success",
  "data": { ... },
  "requestId": "xxx"
}
```

### 错误响应

```json
{
  "code": "Error",
  "message": "错误描述",
  "requestId": "xxx"
}
```

## 错误码

| 错误码 | 描述 |
|--------|------|
| Success | 成功 |
| Error | 通用错误 |
| TaskCreateFailed | 任务创建失败 |
| TaskExecuteFailed | 任务执行失败 |
| TaskTimeout | 任务超时 |
| ResultParseFailed | 结果解析失败 |
| GetResultFailed | 获取结果失败 |

## 权限错误处理

当遇到权限错误时，服务器会返回增强的错误信息：

```
权限不足

这可能是权限问题。请检查：
1. 您的阿里云账号是否具有访问 SysOM 服务的权限
2. 是否正确配置了 AccessKey ID 和 AccessKey Secret
3. 是否需要为 RAM 用户授予 AliyunSysOMFullAccess 权限
4. 如果使用 STS Token，请检查 Token 是否过期
```
