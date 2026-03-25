/*
CLI Chat 命令 - AI 交互式排障

通过 MCP 协议与 AI 诊断引擎交互
*/
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// MCPRequest MCP 请求结构
type MCPRequest struct {
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
}

// MCPResponse MCP 响应结构
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	} `json:"error,omitempty"`
}

func main() {
	// 检查 MCP Server 是否运行
	mcpURL := getMCPURL()
	if !checkMCPServer(mcpURL) {
		fmt.Println("❌ 无法连接到 MCP Server")
		fmt.Printf("请确保 MCP Server 正在运行：%s\n", mcpURL)
		fmt.Println("\n启动命令:")
		fmt.Println("  go run cmd/server/main.go")
		os.Exit(1)
	}

	fmt.Println("🤖 Siriusec AI Assistant")
	fmt.Println("=========================")
	fmt.Println("输入 'help' 查看帮助，'quit' 退出\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			fmt.Println("再见！👋")
			break
		}

		if input == "help" {
			printHelp()
			continue
		}

		// 调用 AI 诊断
		response, err := callSmartDiagnose(mcpURL, input)
		if err != nil {
			fmt.Printf("❌ 调用失败：%v\n\n", err)
			continue
		}

		// 渲染 Markdown 响应
		renderResponse(response)
	}
}

// getMCPURL 获取 MCP Server URL
func getMCPURL() string {
	if url := os.Getenv("MCP_URL"); url != "" {
		return url
	}
	return "http://localhost:7140/mcp/unified"
}

// checkMCPServer 检查 MCP Server 是否运行
func checkMCPServer(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// callSmartDiagnose 调用智能诊断
func callSmartDiagnose(url, query string) (string, error) {
	// 构建 MCP 请求
	reqBody := MCPRequest{
		Method:  "tools/call",
		JSONRPC: "2.0",
		ID:      1,
		Params: map[string]interface{}{
			"name": "smart_diagnose",
			"arguments": map[string]interface{}{
				"query": query,
			},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	// 发送 HTTP 请求
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	if mcpResp.Error != nil {
		return "", fmt.Errorf("MCP error: %s", mcpResp.Error.Message)
	}

	// 解析结果
	var result struct {
		Content []struct {
			Type     string `json:"type"`
			Text     string `json:"text"`
			MimeType string `json:"mimeType,omitempty"`
		} `json:"content"`
	}

	if err := json.Unmarshal(mcpResp.Result, &result); err != nil {
		return "", fmt.Errorf("parse result failed: %w", err)
	}

	// 提取文本内容
	var textContent string
	for _, content := range result.Content {
		if content.Type == "text" {
			textContent += content.Text + "\n"
		}
	}

	return textContent, nil
}

// renderResponse 渲染响应
func renderResponse(content string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println(content)
	fmt.Println(strings.Repeat("=", 60) + "\n")
}

func printHelp() {
	fmt.Println("\n可用命令:")
	fmt.Println("  help          - 显示帮助")
	fmt.Println("  quit/exit     - 退出程序")
	fmt.Println()
	fmt.Println("示例问题:")
	fmt.Println("  - 服务器 CPU 很高，帮我看看")
	fmt.Println("  - 内存使用率多少？")
	fmt.Println("  - 磁盘空间够吗？")
	fmt.Println("  - 为什么系统这么卡？")
	fmt.Println()
}
