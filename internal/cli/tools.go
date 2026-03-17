// Package cli 提供命令行工具功能
package cli

import (
	"encoding/json"
	"fmt"
	"os"
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
	"text/tabwriter"

	"go.uber.org/zap"
)

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
}

// ListTools 列出所有工具
func ListTools(format string) error {
	// 禁用日志输出
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()

	server := mcp.NewServer("cli", "1.0.0")

	// 注册所有工具
	am.RegisterTools(server)
	memdiag.RegisterTools(server)
	iodiag.RegisterTools(server)
	netdiag.RegisterTools(server)
	scheddiag.RegisterTools(server)
	otherdiag.RegisterTools(server)
	crashagent.RegisterTools(server)
	initial.RegisterTools(server)

	tools := server.GetAllTools()

	// 构建工具信息列表
	var toolList []ToolInfo
	for _, tool := range tools {
		category := "unknown"
		if len(tool.Tags) > 0 {
			category = tool.Tags[0]
		}
		toolList = append(toolList, ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			Category:    category,
			Status:      "✓",
		})
	}

	switch format {
	case "json":
		return printToolsJSON(toolList)
	case "table":
		return printToolsTable(toolList)
	default:
		return printToolsSimple(toolList)
	}
}

func printToolsJSON(tools []ToolInfo) error {
	data, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printToolsTable(tools []ToolInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCATEGORY\tDESCRIPTION")
	fmt.Fprintln(w, "----\t--------\t-----------")
	for _, tool := range tools {
		fmt.Fprintf(w, "%s\t%s\t%s\n", tool.Name, tool.Category, tool.Description)
	}
	return w.Flush()
}

func printToolsSimple(tools []ToolInfo) error {
	fmt.Printf("Total tools: %d\n\n", len(tools))

	// 按分类分组
	categories := make(map[string][]ToolInfo)
	for _, tool := range tools {
		categories[tool.Category] = append(categories[tool.Category], tool)
	}

	for category, categoryTools := range categories {
		fmt.Printf("[%s] (%d tools)\n", category, len(categoryTools))
		for _, tool := range categoryTools {
			fmt.Printf("  ✓ %s - %s\n", tool.Name, tool.Description)
		}
		fmt.Println()
	}

	return nil
}
