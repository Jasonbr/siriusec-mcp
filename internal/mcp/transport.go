// Package mcp 传输层实现
// 支持stdio、SSE、streamable-http三种传输模式
package mcp

import (
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

// ServeStreamableHTTP 启动streamable-http模式服务
func ServeStreamableHTTP(mcpServer *server.MCPServer, host string, port int, path string) error {
	// 创建多路复用器
	mux := http.NewServeMux()

	// 使用mcp-go提供的Streamable HTTP服务器
	httpServer := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath(path),
	)

	// MCP 端点
	mux.Handle(path, httpServer)
	mux.Handle(path+"/", httpServer)

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
