package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// 获取 web-ui 目录路径
	webDir := filepath.Join(getCurrentDir(), "web-ui")

	// 检查目录是否存在
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		log.Fatalf("Web UI directory not found: %s", webDir)
	}

	// 创建文件服务器
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// 添加 CORS 支持（用于访问 MCP 服务器）
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 处理健康检查端点
		if r.URL.Path == "/api/health" || r.URL.Path == "/api/mcp/unified" {
			proxyRequest(w, r)
			return
		}

		// 其他 API 请求也代理到 MCP 服务器
		proxyRequest(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🌐 Siriusec MCP Web UI\n")
	fmt.Printf("📍 Server starting on http://localhost%s\n", addr)
	fmt.Printf("📁 Serving files from: %s\n", webDir)
	fmt.Printf("\n按 Ctrl+C 停止服务\n\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func getCurrentDir() string {
	// 使用工作目录而不是可执行文件目录
	wd, err := os.Getwd()
	if err != nil {
		executable, _ := os.Executable()
		return filepath.Dir(executable)
	}
	return wd
}

func proxyRequest(w http.ResponseWriter, r *http.Request) {
	// 简单的代理逻辑，实际使用中可能需要更完善的实现
	client := &http.Client{}

	// 移除 /api 前缀，转换为 MCP 服务器的路径
	targetPath := strings.TrimPrefix(r.URL.Path, "/api")
	if targetPath == "" {
		targetPath = "/"
	}

	// 创建新请求
	req, err := http.NewRequest(r.Method, "http://localhost:7140"+targetPath, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 复制响应状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}
