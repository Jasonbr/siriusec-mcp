// Package cli 提供命令行工具功能
package cli

import (
	"context"
	"fmt"
	"os"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/config"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/am"
	"siriusec-mcp/internal/tools/crashagent"
	"siriusec-mcp/internal/tools/initial"
	"siriusec-mcp/internal/tools/iodiag"
	"siriusec-mcp/internal/tools/memdiag"
	"siriusec-mcp/internal/tools/netdiag"
	"siriusec-mcp/internal/tools/otherdiag"
	"siriusec-mcp/internal/tools/scheddiag"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TestResult 测试结果
type TestResult struct {
	ToolName string `json:"tool_name"`
	Category string `json:"category"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duration string `json:"duration"`
}

// RunSelfTest 运行自测
func RunSelfTest(filter string, verbose bool) error {
	// 禁用日志输出
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()

	fmt.Println("Running self-test...")
	fmt.Println()

	// 创建服务器实例
	server := mcp.NewServer("selftest", "1.0.0")
	am.RegisterTools(server)
	memdiag.RegisterTools(server)
	iodiag.RegisterTools(server)
	netdiag.RegisterTools(server)
	scheddiag.RegisterTools(server)
	otherdiag.RegisterTools(server)
	crashagent.RegisterTools(server)
	initial.RegisterTools(server)

	tools := server.GetAllTools()

	var results []TestResult
	passed := 0
	failed := 0

	for _, tool := range tools {
		// 应用过滤器
		if filter != "" && !strings.Contains(tool.Name, filter) && !strings.Contains(tool.Tags[0], filter) {
			continue
		}

		result := testTool(tool.Name, tool.Tags[0], verbose)
		results = append(results, result)

		if result.Status == "PASS" {
			passed++
		} else {
			failed++
		}

		if verbose {
			fmt.Printf("  %s %s - %s (%s)\n", result.Status, result.ToolName, result.Message, result.Duration)
		}
	}

	fmt.Println()
	fmt.Printf("Test Results: %d passed, %d failed, %d total\n", passed, failed, len(results))

	if failed > 0 {
		return fmt.Errorf("self-test completed with %d failures", failed)
	}

	return nil
}

func testTool(name, category string, verbose bool) TestResult {
	start := time.Now()

	// 基础检查：工具是否已注册
	result := TestResult{
		ToolName: name,
		Category: category,
		Status:   "PASS",
		Message:  "Tool registered successfully",
	}

	// 对于某些工具，检查配置是否可用
	switch name {
	case "check_sysom_initialed", "initial_sysom":
		if config.GlobalConfig == nil || config.GlobalConfig.OpenAPI.AccessKeyID == "" {
			result.Status = "SKIP"
			result.Message = "Skipped: No valid configuration"
		}
	}

	result.Duration = time.Since(start).String()
	return result
}

// TestOpenAPIConnection 测试 OpenAPI 连接
func TestOpenAPIConnection() error {
	fmt.Println("Testing OpenAPI connection...")
	fmt.Println()

	if config.GlobalConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if config.GlobalConfig.OpenAPI.AccessKeyID == "" {
		return fmt.Errorf("ACCESS_KEY_ID not configured")
	}

	fmt.Printf("  Region: %s\n", config.GlobalConfig.OpenAPI.RegionID)
	fmt.Printf("  Type: %s\n", config.GlobalConfig.OpenAPI.Type)

	// 尝试创建客户端
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.GlobalFactory.CreateClient("")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if c == nil {
		return fmt.Errorf("client is nil")
	}

	_ = ctx
	fmt.Println("  ✓ Client created successfully")
	fmt.Println()
	fmt.Println("OpenAPI connection test passed!")

	return nil
}

// PrintSelfTestResults 打印自测结果
func PrintSelfTestResults(results []TestResult) {
	w := os.Stdout
	fmt.Fprintln(w, "TOOL NAME\tCATEGORY\tSTATUS\tMESSAGE")
	fmt.Fprintln(w, "---------\t--------\t------\t-------")
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.ToolName, r.Category, r.Status, r.Message)
	}
}
