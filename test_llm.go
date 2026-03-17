package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	baseURL := "http://127.0.0.1:7140"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	fmt.Println("==========================================")
	fmt.Println("测试阿里百炼 kimi-k2.5 模型")
	fmt.Println("==========================================")
	fmt.Printf("服务器地址: %s\n\n", baseURL)

	client := &http.Client{Timeout: 120 * time.Second}

	// 1. 测试健康检查
	fmt.Println("1. 测试健康检查...")
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("   ✗ 失败: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("   ✓ 健康检查通过 (HTTP %d)\n\n", resp.StatusCode)

	// 2. 初始化 MCP 会话
	fmt.Println("2. 初始化 MCP 会话...")
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	initBody, _ := json.Marshal(initReq)
	resp, err = client.Post(baseURL+"/mcp/unified", "application/json", bytes.NewReader(initBody))
	if err != nil {
		fmt.Printf("   ✗ 失败: %v\n", err)
		return
	}
	initResp, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("   ✓ 初始化成功\n")
	fmt.Printf("   响应: %s\n\n", string(initResp))

	// 3. 测试 smart_diagnose 工具
	fmt.Println("3. 测试 smart_diagnose 工具...")
	toolReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "smart_diagnose",
			"arguments": map[string]interface{}{
				"symptom": "服务器 CPU 使用率突然飙升到 100%",
				"context": "阿里云 ECS 实例，运行 Java 应用",
			},
		},
	}

	toolBody, _ := json.Marshal(toolReq)
	fmt.Println("   发送请求到百炼模型...")
	start := time.Now()
	resp, err = client.Post(baseURL+"/mcp/unified", "application/json", bytes.NewReader(toolBody))
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
		return
	}
	toolResp, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	duration := time.Since(start)

	// 解析响应
	var result map[string]interface{}
	json.Unmarshal(toolResp, &result)

	if result["error"] != nil {
		fmt.Printf("   ✗ 工具调用失败: %v\n", result["error"])
	} else {
		fmt.Printf("   ✓ 工具调用成功 (耗时: %v)\n", duration)
		fmt.Printf("   原始响应: %s\n\n", string(toolResp))

		// 尝试解析结果
		if resultMap, ok := result["result"].(map[string]interface{}); ok {
			if content, ok := resultMap["content"].([]interface{}); ok && len(content) > 0 {
				if contentMap, ok := content[0].(map[string]interface{}); ok {
					if text, ok := contentMap["text"].(string); ok {
						fmt.Printf("========== AI 诊断结果 ==========\n")
						fmt.Println(text)
						fmt.Println("===================================\n")
					} else {
						fmt.Printf("内容格式: %v\n", contentMap)
					}
				}
			} else {
				fmt.Printf("结果内容: %v\n", resultMap)
			}
		} else {
			fmt.Printf("无法解析结果: %v\n", result["result"])
		}
	}

	fmt.Println("测试完成!")
}
