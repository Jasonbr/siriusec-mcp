/*
其他诊断工具

提供vmcore_analysis、disk_analysis等其他诊断功能
*/
package otherdiag

import (
	"context"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/diagnosis"
	"siriusec-mcp/pkg/models"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册其他诊断工具
func RegisterTools(s *mcp.Server) {
	// vmcore_analysis (原名vmcore)
	s.RegisterTool("vmcore_analysis",
		mcpp.NewTool("vmcore_analysis",
			mcpp.WithDescription("宕机诊断工具，分析操作系统崩溃的原因，通过分析内核panic产生的core dump文件"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
		),
		handleVmcoreAnalysis,
		"sysom_otherdiag",
	)

	// disk_analysis (原名diskanalysis)
	s.RegisterTool("disk_analysis",
		mcpp.NewTool("disk_analysis",
			mcpp.WithDescription("磁盘分析诊断工具，分析系统中磁盘的使用情况"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
		),
		handleDiskAnalysis,
		"sysom_otherdiag",
	)
}

// handleVmcoreAnalysis 处理vmcore_analysis诊断请求
func handleVmcoreAnalysis(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleOtherDiagnosis(ctx, request, "vmcore")
}

// handleDiskAnalysis 处理disk_analysis诊断请求
func handleDiskAnalysis(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleOtherDiagnosis(ctx, request, "diskanalysis")
}

// handleOtherDiagnosis 通用其他诊断处理
func handleOtherDiagnosis(ctx context.Context, request mcpp.CallToolRequest, serviceName string) (*mcpp.CallToolResult, error) {
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
