// Package mcp 传输层实现
// 支持stdio、SSE、streamable-http三种传输模式
package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/version"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// ServeSSE 启动SSE模式服务
func ServeSSE(mcpServer *server.MCPServer, host string, port int, path string) error {
	// 使用mcp-go提供的SSE服务器
	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithBasePath(path),
		server.WithSSEEndpoint(path),
		server.WithMessageEndpoint(path+"/message"),
	)

	addr := fmt.Sprintf("%s:%d", host, port)
	logger.Infof("SSE server listening on %s%s", addr, path)
	return sseServer.Start(addr)
}

// ServeStreamableHTTP 启动 streamable-http 模式服务
func ServeStreamableHTTP(mcpServer *server.MCPServer, host string, port int, path string) error {
	// 创建多路复用器
	mux := http.NewServeMux()

	// 使用 mcp-go 提供的 Streamable HTTP 服务器
	httpServer := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath(path),
	)

	// MCP 端点 - 使用自定义的 Handler 包装（添加 CORS 和 Session ID 注入）
	mux.Handle(path, withCORS(withSessionIDInjector(httpServer)))
	mux.Handle(path+"/", withCORS(withSessionIDInjector(httpServer)))

	// 健康检查端点
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/version", versionHandler)
	mux.HandleFunc("/ready", readinessHandler)

	addr := fmt.Sprintf("%s:%d", host, port)
	logger.Info("Streamable HTTP server starting",
		zap.String("addr", addr),
		zap.String("path", path))

	return http.ListenAndServe(addr, mux)
}

// withCORS 添加 CORS 支持的中间件
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 允许所有来源（开发环境）
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")

		// 处理 OPTIONS 预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// withSessionIDInjector 自定义 HTTP Handler，用于在响应中注入 Session ID
func withSessionIDInjector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 创建响应捕获器
		recorder := &responseRecorder{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		// 调用下一个 handler
		next.ServeHTTP(recorder, r)

		// 从响应头获取 Session ID
		sessionID := recorder.Header().Get("Mcp-Session-Id")
		if sessionID != "" {
			// 读取响应体
			var response map[string]interface{}
			if err := json.Unmarshal(recorder.body.Bytes(), &response); err == nil {
				// 在响应体中添加_sessionId 字段
				if result, ok := response["result"].(map[string]interface{}); ok {
					result["_sessionId"] = sessionID
				}

				// 重新编码响应
				newBody, _ := json.Marshal(response)
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
				w.Write(newBody)
				return
			}
		}

		// 如果没有 Session ID 或解析失败，使用原始响应
		w.Write(recorder.body.Bytes())
	})
}

// responseRecorder 用于捕获响应的自定义 ResponseWriter
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	written    bool
	statusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.statusCode = http.StatusOK
	}
	r.body.Write(b)
	r.written = true
	return len(b), nil
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.Header().Set("Content-Type", "application/json")
}

// healthHandler 健康检查处理
func healthHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := map[string]interface{}{
		"status":    "healthy",
		"version":   version.Short(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"system": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"memory": map[string]interface{}{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"gc_count":    m.NumGC,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// versionHandler 版本信息处理
func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(version.Full())
}

// readinessHandler 就绪检查处理
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
