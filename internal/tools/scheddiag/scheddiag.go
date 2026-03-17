/*
调度诊断工具

提供delay、loadtask等调度诊断功能
*/
package scheddiag

import (
	"context"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/diagnosis"
	"siriusec-mcp/pkg/models"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册调度诊断工具
func RegisterTools(s *mcp.Server) {
	// delay
	s.RegisterTool("delay",
		mcpp.NewTool("delay",
			mcpp.WithDescription("调度抖动诊断工具，分析CPU长时间不进行任务切换导致的问题"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
			mcpp.WithString("duration", mcpp.Description("诊断持续的时长(s)"), mcpp.DefaultString("20")),
			mcpp.WithString("threshold", mcpp.Description("判定出现抖动的阈值(ms)"), mcpp.DefaultString("20")),
		),
		handleDelay,
		"sysom_scheddiag",
	)

	// loadtask
	s.RegisterTool("loadtask",
		mcpp.NewTool("loadtask",
			mcpp.WithDescription("系统负载诊断工具，分析系统负载异常原因及详细信息"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
		),
		handleLoadTask,
		"sysom_scheddiag",
	)
}

// handleDelay 处理delay诊断请求
func handleDelay(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleSchedDiagnosis(ctx, request, "delay")
}

// handleLoadTask 处理loadtask诊断请求
func handleLoadTask(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleSchedDiagnosis(ctx, request, "loadtask")
}

// handleSchedDiagnosis 通用调度诊断处理
func handleSchedDiagnosis(ctx context.Context, request mcpp.CallToolRequest, serviceName string) (*mcpp.CallToolResult, error) {
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
	if val, ok := mcp.GetStringParam(request, "duration"); ok {
		params["duration"] = val
	}
	if val, ok := mcp.GetStringParam(request, "threshold"); ok {
		params["threshold"] = val
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
