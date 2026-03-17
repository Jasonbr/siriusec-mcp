/*
网络诊断工具

提供packetdrop、netjitter等网络诊断功能
*/
package netdiag

import (
	"context"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/diagnosis"
	"siriusec-mcp/pkg/models"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册网络诊断工具
func RegisterTools(s *mcp.Server) {
	// packetdrop
	s.RegisterTool("packetdrop",
		mcpp.NewTool("packetdrop",
			mcpp.WithDescription("网络丢包诊断工具，分析数据包通过网络传输过程中的丢失问题"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
		),
		handlePacketDrop,
		"sysom_netdiagnose",
	)

	// netjitter
	s.RegisterTool("netjitter",
		mcpp.NewTool("netjitter",
			mcpp.WithDescription("网络抖动诊断工具，分析数据包在网络传输过程中的不稳定现象"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("实例地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，仅支持ecs")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
			mcpp.WithString("duration", mcpp.Description("诊断持续的时长(s)"), mcpp.DefaultString("20")),
			mcpp.WithString("threshold", mcpp.Description("判定出现抖动的阈值(ms)"), mcpp.DefaultString("10")),
		),
		handleNetJitter,
		"sysom_netdiagnose",
	)
}

// handlePacketDrop 处理packetdrop诊断请求
func handlePacketDrop(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleNetDiagnosis(ctx, request, "packetdrop")
}

// handleNetJitter 处理netjitter诊断请求
func handleNetJitter(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleNetDiagnosis(ctx, request, "netjitter")
}

// handleNetDiagnosis 通用网络诊断处理
func handleNetDiagnosis(ctx context.Context, request mcpp.CallToolRequest, serviceName string) (*mcpp.CallToolResult, error) {
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
