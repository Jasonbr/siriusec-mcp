/*
内存诊断工具

提供memgraph、javamem、oomcheck等内存诊断功能
*/
package memdiag

import (
	"context"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/diagnosis"
	"siriusec-mcp/pkg/models"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册内存诊断工具
func RegisterTools(s *mcp.Server) {
	// memgraph
	s.RegisterTool("memgraph",
		mcpp.NewTool("memgraph",
			mcpp.WithDescription("内存全景分析工具"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，可选值：ecs，auto")),
			mcpp.WithString("instance", mcpp.Description("实例ID")),
			mcpp.WithString("pod", mcpp.Description("Pod名称")),
			mcpp.WithString("clusterType", mcpp.Description("集群类型")),
			mcpp.WithString("clusterId", mcpp.Description("集群ID")),
			mcpp.WithString("namespace", mcpp.Description("Pod命名空间")),
		),
		handleMemGraph,
		"sysom_memdiag",
	)

	// javamem
	s.RegisterTool("javamem",
		mcpp.NewTool("javamem",
			mcpp.WithDescription("Java内存诊断工具"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，可选值：ecs，auto")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
			mcpp.WithString("pid", mcpp.Description("Java进程Pid")),
			mcpp.WithString("pod", mcpp.Description("Pod名称")),
			mcpp.WithString("duration", mcpp.Description("JNI内存分配profiling时长"), mcpp.DefaultString("0")),
			mcpp.WithString("clusterType", mcpp.Description("集群类型")),
			mcpp.WithString("clusterId", mcpp.Description("集群ID")),
			mcpp.WithString("namespace", mcpp.Description("Pod命名空间")),
		),
		handleJavaMem,
		"sysom_memdiag",
	)

	// oomcheck
	s.RegisterTool("oomcheck",
		mcpp.NewTool("oomcheck",
			mcpp.WithDescription("OOM诊断工具"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Required(), mcpp.Description("地域")),
			mcpp.WithString("channel", mcpp.Required(), mcpp.Description("诊断通道，可选值：ecs，auto")),
			mcpp.WithString("instance", mcpp.Description("实例ID")),
			mcpp.WithString("pod", mcpp.Description("Pod名称")),
			mcpp.WithString("time", mcpp.Description("时间戳")),
			mcpp.WithString("clusterType", mcpp.Description("集群类型")),
			mcpp.WithString("clusterId", mcpp.Description("集群ID")),
			mcpp.WithString("namespace", mcpp.Description("Pod命名空间")),
		),
		handleOOMCheck,
		"sysom_memdiag",
	)
}

// handleMemGraph 处理memgraph诊断请求
func handleMemGraph(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleDiagnosis(ctx, request, "memgraph")
}

// handleJavaMem 处理javamem诊断请求
func handleJavaMem(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleDiagnosis(ctx, request, "javamem")
}

// handleOOMCheck 处理oomcheck诊断请求
func handleOOMCheck(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	return handleDiagnosis(ctx, request, "oomcheck")
}

// handleDiagnosis 通用诊断处理
func handleDiagnosis(ctx context.Context, request mcpp.CallToolRequest, serviceName string) (*mcpp.CallToolResult, error) {
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
	if val, ok := mcp.GetStringParam(request, "pod"); ok {
		params["pod"] = val
	}
	if val, ok := mcp.GetStringParam(request, "clusterType"); ok {
		params["clusterType"] = val
	}
	if val, ok := mcp.GetStringParam(request, "clusterId"); ok {
		params["clusterId"] = val
	}
	if val, ok := mcp.GetStringParam(request, "namespace"); ok {
		params["namespace"] = val
	}
	if val, ok := mcp.GetStringParam(request, "pid"); ok {
		params["Pid"] = val
	}
	if val, ok := mcp.GetStringParam(request, "duration"); ok {
		params["duration"] = val
	}
	if val, ok := mcp.GetStringParam(request, "time"); ok {
		params["time"] = val
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
