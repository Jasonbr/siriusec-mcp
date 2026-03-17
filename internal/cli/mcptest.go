// Package cli 提供命令行工具功能
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MCPTestResult MCP 测试结果
type MCPTestResult struct {
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duration string `json:"duration"`
}

// RunMCPTest 运行 MCP 连接测试
func RunMCPTest(baseURL string) []MCPTestResult {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:7140"
	}

	fmt.Printf("Testing MCP server at %s\n", baseURL)
	fmt.Println()

	var results []MCPTestResult

	// 测试健康检查端点
	results = append(results, testEndpoint(
		baseURL+"/health",
		"GET",
		nil,
		"Health check",
	))

	// 测试版本端点
	results = append(results, testEndpoint(
		baseURL+"/version",
		"GET",
		nil,
		"Version endpoint",
	))

	// 测试就绪端点
	results = append(results, testEndpoint(
		baseURL+"/ready",
		"GET",
		nil,
		"Readiness check",
	))

	// 测试 MCP initialize
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "mcptest",
				"version": "1.0.0",
			},
		},
	}
	results = append(results, testEndpoint(
		baseURL+"/mcp/unified",
		"POST",
		initReq,
		"MCP initialize",
	))

	return results
}

func testEndpoint(url, method string, body interface{}, name string) MCPTestResult {
	start := time.Now()
	result := MCPTestResult{
		Endpoint: name,
		Status:   "FAIL",
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			result.Message = fmt.Sprintf("Failed to marshal request: %v", err)
			result.Duration = time.Since(start).String()
			return result
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Message = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}
	defer resp.Body.Close()

	result.Duration = time.Since(start).String()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = "PASS"
		result.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result
}

// PrintMCPTestResults 打印 MCP 测试结果
func PrintMCPTestResults(results []MCPTestResult) {
	fmt.Println()
	fmt.Println("MCP Test Results:")
	fmt.Println("-----------------")

	passed := 0
	failed := 0

	for _, r := range results {
		status := "✓"
		if r.Status != "PASS" {
			status = "✗"
		}
		fmt.Printf("  %s %-30s %s (%s)\n", status, r.Endpoint, r.Message, r.Duration)

		if r.Status == "PASS" {
			passed++
		} else {
			failed++
		}
	}

	fmt.Println()
	fmt.Printf("Results: %d passed, %d failed\n", passed, failed)
}
