# Siriusec_MCP Go Server Dockerfile
# 多阶段构建，减小最终镜像体积

# 阶段1：构建阶段
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
# 使用静态链接，确保在alpine中正常运行
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/siriusec-mcp \
    cmd/server/main.go

# 阶段2：运行阶段
FROM alpine:3.19

# 安装ca证书和时区数据
RUN apk add --no-cache ca-certificates tzdata

# 创建非root用户运行应用
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/siriusec-mcp /app/siriusec-mcp

# 更改文件所有者
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口（SSE和streamable-http模式使用）
EXPOSE 7140

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7140/health || exit 1

# 默认以streamable-http模式运行
ENTRYPOINT ["/app/siriusec-mcp"]
CMD ["run", "--streamable-http", "--host", "0.0.0.0", "--port", "7140"]
