/*
IO诊断工具

提供iofsstat、iodiagnose等IO诊断功能
*/
package iodiag

import (
	"context"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/diagnosis"
	"siriusec-mcp/pkg/models"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册IO诊断工具
func RegisterTools(s *mcp.Server) {
	// iofsstat
	s.RegisterTool("iofsstat",
		mcpp.NewTool("iofsstat",
			mcpp.WithDescription("IO流量分析工具，分析系统中IO流量的归属"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
			mcpp.WithString("timeout", mcpp.Description("诊断时长"), mcpp.DefaultString("15")),
			mcpp.WithString("disk", mcpp.Description("磁盘名称，例如sda等，缺省为所有磁盘")),
		),
		handleIOFSStat,
		"sysom_iodiag",
	)

	// iodiagnose
	s.RegisterTool("iodiagnose",
		mcpp.NewTool("iodiagnose",
			mcpp.WithDescription("IO一键诊断工具，专注于IO高延迟、IO Burst及IO Wait等问题"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
			mcpp.WithString("timeout", mcpp.Description("诊断时长，默认为30秒，不建议低于30秒"), mcpp.DefaultString("30")),
		),
		handleIODiagnose,
		"sysom_iodiag",
	)
}

// handleIOFSStat 处理iofsstat诊断请求
func handleIOFSStat(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleIODiagnosis(ctx, request, "iofsstat")
}

// handleIODiagnose 处理iodiagnose诊断请求
func handleIODiagnose(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleIODiagnosis(ctx, request, "iodiagnose")
}

// handleIODiagnosis 通用IO诊断处理
func handleIODiagnosis(ctx context.Context, request mcpp.CallToolRequest, serviceName string) (*mcpp.CallToolResult, error) {
	uid, _ := mcp.GetStringParam(request, "uid")
	region, _ := mcp.GetStringParam(request, "region")
	channel, _ := mcp.GetStringParam(request, "channel")

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 构建参数
	params := make(map[string]interface{})
	if val, ok := mcp.GetStringParam(request, "instance"); ok {
		params["instance"] = val
	}
	if val, ok := mcp.GetStringParam(request, "timeout"); ok {
		params["timeout"] = val
	}
	if val, ok := mcp.GetStringParam(request, "disk"); ok {
		params["disk"] = val
	}
	params["region"] = region

	// 创建诊断请求
	req := &diagnosis.Request{
		ServiceName: serviceName,
		Channel:     channel,
		Region:      region,
		Params:      params,
	}

	// 执行诊断
	helper := diagnosis.NewHelper(c, 0, 0)
	resp := helper.Execute(ctx, req)

	// 检查权限错误
	if resp.Code != diagnosis.ResultCodeSuccess && models.IsPermissionError(resp.Message) {
		resp.Message = models.EnhancePermissionErrorMessage(resp.Message)
	}

	// 转换为MCP响应
	mcpResp := resp.ToMCPResponse()
	if resp.Code == diagnosis.ResultCodeSuccess {
		return mcp.CreateSuccessResult(mcpResp)
	}
	return mcp.CreateErrorResult(resp.Message)
}
