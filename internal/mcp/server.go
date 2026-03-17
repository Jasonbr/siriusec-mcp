/*
MCP服务器核心

实现统一的MCP服务器，支持工具注册和管理
*/
package mcp

import (
	"encoding/json"
	"fmt"
	"siriusec-mcp/internal/logger"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server MCP服务器
type Server struct {
	mcpServer *server.MCPServer
	tools     map[string]*ToolInfo
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string
	Description string
	Tags        []string
	Handler     ToolHandler
	Schema      *mcp.Tool
}

// ToolHandler 工具处理函数类型
type ToolHandler = server.ToolHandlerFunc

// NewServer 创建MCP服务器
func NewServer(name string, version string) *Server {
	s := &Server{
		mcpServer: server.NewMCPServer(
			name,
			version,
		),
		tools: make(map[string]*ToolInfo),
	}

	return s
}

// RegisterTool 注册工具
func (s *Server) RegisterTool(name string, tool mcp.Tool, handler ToolHandler, tags ...string) {
	logger.Infof("Registering tool: %s", name)

	// 保存工具信息
	s.tools[name] = &ToolInfo{
		Name:        name,
		Description: tool.Description,
		Tags:        tags,
		Handler:     handler,
		Schema:      &tool,
	}

	// 注册到MCP服务器
	s.mcpServer.AddTool(tool, handler)
}

// GetTool 获取工具信息
func (s *Server) GetTool(name string) *ToolInfo {
	return s.tools[name]
}

// GetToolsByTag 根据标签获取工具
func (s *Server) GetToolsByTag(tag string) []*ToolInfo {
	var result []*ToolInfo
	for _, tool := range s.tools {
		for _, t := range tool.Tags {
			if t == tag {
				result = append(result, tool)
				break
			}
		}
	}
	return result
}

// GetAllTools 获取所有工具
func (s *Server) GetAllTools() []*ToolInfo {
	result := make([]*ToolInfo, 0, len(s.tools))
	for _, tool := range s.tools {
		result = append(result, tool)
	}
	return result
}

// GetMCPServer 获取底层MCP服务器
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

// ServeStdio 启动stdio模式服务
func (s *Server) ServeStdio() error {
	logger.Info("Starting MCP server in stdio mode")
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE 启动SSE模式服务
func (s *Server) ServeSSE(host string, port int, path string) error {
	logger.Infof("Starting MCP server in SSE mode on %s:%d%s", host, port, path)
	return ServeSSE(s.mcpServer, host, port, path)
}

// ServeStreamableHTTP 启动streamable-http模式服务
func (s *Server) ServeStreamableHTTP(host string, port int, path string) error {
	logger.Infof("Starting MCP server in streamable-http mode on %s:%d%s", host, port, path)
	return ServeStreamableHTTP(s.mcpServer, host, port, path)
}

// CreateSuccessResult 创建成功的工具调用结果
func CreateSuccessResult(data interface{}) (*mcp.CallToolResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// CreateErrorResult 创建错误的工具调用结果
func CreateErrorResult(message string) (*mcp.CallToolResult, error) {
	result := map[string]interface{}{
		"code":    "Error",
		"message": message,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal error result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		IsError: true,
	}, nil
}

// GetStringParam 获取字符串参数
func GetStringParam(request mcp.CallToolRequest, name string) (string, bool) {
	val := request.GetString(name, "")
	if val != "" {
		return val, true
	}
	return "", false
}

// GetIntParam 获取整数参数
func GetIntParam(request mcp.CallToolRequest, name string) (int, bool) {
	val := request.GetInt(name, 0)
	if val != 0 {
		return val, true
	}
	return 0, false
}

// GetBoolParam 获取布尔参数
func GetBoolParam(request mcp.CallToolRequest, name string) (bool, bool) {
	val := request.GetBool(name, false)
	return val, true
}

// GetFloat64Param 获取浮点数参数
func GetFloat64Param(request mcp.CallToolRequest, name string) (float64, bool) {
	val := request.GetFloat(name, 0)
	if val != 0 {
		return val, true
	}
	return 0, false
}
