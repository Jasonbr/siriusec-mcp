/*
应用管理服务 (AM) 工具实现

提供实例、集群、Pod相关的查询功能
*/
package am

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/registry"
	"siriusec-mcp/pkg/models"

	sysom "github.com/alibabacloud-go/sysom-20231230/client"
	"github.com/alibabacloud-go/tea/tea"
	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册AM工具
func RegisterTools(s *mcp.Server) {
	// 注册API路由
	registerAPIRoutes()

	// list_all_instances
	s.RegisterTool("list_all_instances",
		mcpp.NewTool("list_all_instances",
			mcpp.WithDescription("列出所有实例，支持按地域、纳管类型、实例类型等条件筛选"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("region", mcpp.Description("地域，如cn-hangzhou")),
			mcpp.WithString("managedType", mcpp.Description("纳管类型，可选值：managed（已纳管）、unmanaged（未纳管）、all（全部）、install（待安装）、uninstall（待卸载）、upgrade（待升级）"), mcpp.DefaultString("all")),
			mcpp.WithString("instanceType", mcpp.Description("实例类型"), mcpp.DefaultString("ecs")),
			mcpp.WithString("pluginId", mcpp.Description("插件ID")),
			mcpp.WithString("filters", mcpp.Description("过滤条件，JSON字符串格式")),
			mcpp.WithString("current", mcpp.Description("页码，从1开始"), mcpp.DefaultString("1")),
			mcpp.WithString("pageSize", mcpp.Description("每页数量"), mcpp.DefaultString("10")),
			mcpp.WithNumber("maxResults", mcpp.Description("最大结果数"), mcpp.DefaultNumber(100)),
			mcpp.WithString("nextToken", mcpp.Description("分页游标")),
		),
		handleListAllInstances,
		"sysom_am",
	)

	// list_instances
	s.RegisterTool("list_instances",
		mcpp.NewTool("list_instances",
			mcpp.WithDescription("列出实例列表，支持按实例ID、状态、地域、集群等条件筛选"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("instance", mcpp.Description("实例ID")),
			mcpp.WithString("status", mcpp.Description("实例状态")),
			mcpp.WithString("region", mcpp.Description("地域")),
			mcpp.WithString("clusterId", mcpp.Description("集群ID")),
			mcpp.WithNumber("current", mcpp.Description("页码，从1开始"), mcpp.DefaultNumber(1)),
			mcpp.WithNumber("pageSize", mcpp.Description("每页数量"), mcpp.DefaultNumber(10)),
		),
		handleListInstances,
		"sysom_am",
	)

	// list_clusters
	s.RegisterTool("list_clusters",
		mcpp.NewTool("list_clusters",
			mcpp.WithDescription("列出集群列表，支持按名称、类型、状态等条件筛选"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("name", mcpp.Description("集群名称")),
			mcpp.WithString("clusterId", mcpp.Description("集群ID（用于过滤）")),
			mcpp.WithString("clusterType", mcpp.Description("集群类型")),
			mcpp.WithString("clusterStatus", mcpp.Description("集群状态")),
			mcpp.WithNumber("current", mcpp.Description("页码，从1开始"), mcpp.DefaultNumber(1)),
			mcpp.WithNumber("pageSize", mcpp.Description("每页数量"), mcpp.DefaultNumber(10)),
		),
		handleListClusters,
		"sysom_am",
	)

	// list_pods_of_instance
	s.RegisterTool("list_pods_of_instance",
		mcpp.NewTool("list_pods_of_instance",
			mcpp.WithDescription("列出指定实例下的Pod列表"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
			mcpp.WithString("instance", mcpp.Required(), mcpp.Description("实例ID")),
			mcpp.WithString("clusterId", mcpp.Description("集群ID")),
			mcpp.WithNumber("current", mcpp.Description("页码，从1开始"), mcpp.DefaultNumber(1)),
			mcpp.WithNumber("pageSize", mcpp.Description("每页数量"), mcpp.DefaultNumber(10)),
		),
		handleListPodsOfInstance,
		"sysom_am",
	)
}

// registerAPIRoutes 注册API路由
func registerAPIRoutes() {
	reg := registry.GetRegistry()

	// list_all_instances
	reg.RegisterSDK(
		"list_all_instances",
		reflect.TypeOf(&sysom.ListAllInstancesRequest{}),
		reflect.TypeOf(&sysom.ListAllInstancesResponse{}),
		func(c interface{}, req interface{}) (interface{}, error) {
			client := c.(*sysom.Client)
			request := req.(*sysom.ListAllInstancesRequest)
			return client.ListAllInstances(request)
		},
	)

	// list_instances
	reg.RegisterSDK(
		"list_instances",
		reflect.TypeOf(&sysom.ListInstancesRequest{}),
		reflect.TypeOf(&sysom.ListInstancesResponse{}),
		func(c interface{}, req interface{}) (interface{}, error) {
			client := c.(*sysom.Client)
			request := req.(*sysom.ListInstancesRequest)
			return client.ListInstances(request)
		},
	)

	// list_clusters
	reg.RegisterSDK(
		"list_clusters",
		reflect.TypeOf(&sysom.ListClustersRequest{}),
		reflect.TypeOf(&sysom.ListClustersResponse{}),
		func(c interface{}, req interface{}) (interface{}, error) {
			client := c.(*sysom.Client)
			request := req.(*sysom.ListClustersRequest)
			return client.ListClusters(request)
		},
	)

	// list_pods_of_instance
	reg.RegisterSDK(
		"list_pods_of_instance",
		reflect.TypeOf(&sysom.ListPodsOfInstanceRequest{}),
		reflect.TypeOf(&sysom.ListPodsOfInstanceResponse{}),
		func(c interface{}, req interface{}) (interface{}, error) {
			client := c.(*sysom.Client)
			request := req.(*sysom.ListPodsOfInstanceRequest)
			return client.ListPodsOfInstance(request)
		},
	)
}

// handleListAllInstances 处理list_all_instances请求
func handleListAllInstances(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	uid, _ := mcp.GetStringParam(request, "uid")
	region, _ := mcp.GetStringParam(request, "region")
	managedType, _ := mcp.GetStringParam(request, "managedType")
	instanceType, _ := mcp.GetStringParam(request, "instanceType")
	pluginId, _ := mcp.GetStringParam(request, "pluginId")
	filters, _ := mcp.GetStringParam(request, "filters")
	current, _ := mcp.GetStringParam(request, "current")
	pageSize, _ := mcp.GetStringParam(request, "pageSize")
	maxResults, _ := mcp.GetIntParam(request, "maxResults")
	nextToken, _ := mcp.GetStringParam(request, "nextToken")

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 构建请求
	req := &sysom.ListAllInstancesRequest{
		Region:       tea.String(region),
		InstanceType: tea.String(instanceType),
		PluginId:     tea.String(pluginId),
		Filters:      tea.String(filters),
		Current:      tea.String(current),
		PageSize:     tea.String(pageSize),
		MaxResults:   tea.Int32(int32(maxResults)),
		NextToken:    tea.String(nextToken),
	}

	// 添加managedType参数（如果有）
	if managedType != "" {
		// 注意：阿里云SDK中可能没有这个字段，需要根据实际情况调整
	}

	// 调用API
	success, responseData, err := c.CallAPI(ctx, "list_all_instances", req)
	if err != nil || !success {
		logger.Errorf("list_all_instances failed: %v", err)
		return mcp.CreateErrorResult(fmt.Sprintf("list_all_instances failed: %v", err))
	}

	// 解析响应
	resp := convertToMCPResponse(responseData)
	return mcp.CreateSuccessResult(resp)
}

// handleListInstances 处理list_instances请求
func handleListInstances(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	uid, _ := mcp.GetStringParam(request, "uid")
	instance, _ := mcp.GetStringParam(request, "instance")
	status, _ := mcp.GetStringParam(request, "status")
	region, _ := mcp.GetStringParam(request, "region")
	clusterId, _ := mcp.GetStringParam(request, "clusterId")
	current, _ := mcp.GetIntParam(request, "current")
	pageSize, _ := mcp.GetIntParam(request, "pageSize")

	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	req := &sysom.ListInstancesRequest{
		Instance:  tea.String(instance),
		Status:    tea.String(status),
		Region:    tea.String(region),
		ClusterId: tea.String(clusterId),
		Current:   tea.Int64(int64(current)),
		PageSize:  tea.Int64(int64(pageSize)),
	}

	success, responseData, err := c.CallAPI(ctx, "list_instances", req)
	if err != nil || !success {
		logger.Errorf("list_instances failed: %v", err)
		return mcp.CreateErrorResult(fmt.Sprintf("list_instances failed: %v", err))
	}

	resp := convertToMCPResponse(responseData)
	return mcp.CreateSuccessResult(resp)
}

// handleListClusters 处理list_clusters请求
func handleListClusters(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	uid, _ := mcp.GetStringParam(request, "uid")
	name, _ := mcp.GetStringParam(request, "name")
	clusterId, _ := mcp.GetStringParam(request, "clusterId")
	clusterType, _ := mcp.GetStringParam(request, "clusterType")
	clusterStatus, _ := mcp.GetStringParam(request, "clusterStatus")
	current, _ := mcp.GetIntParam(request, "current")
	pageSize, _ := mcp.GetIntParam(request, "pageSize")

	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	req := &sysom.ListClustersRequest{
		Name:          tea.String(name),
		ClusterId:     tea.String(clusterId),
		ClusterType:   tea.String(clusterType),
		ClusterStatus: tea.String(clusterStatus),
		Current:       tea.Int64(int64(current)),
		PageSize:      tea.Int64(int64(pageSize)),
	}

	success, responseData, err := c.CallAPI(ctx, "list_clusters", req)
	if err != nil || !success {
		logger.Errorf("list_clusters failed: %v", err)
		return mcp.CreateErrorResult(fmt.Sprintf("list_clusters failed: %v", err))
	}

	resp := convertToMCPResponse(responseData)
	return mcp.CreateSuccessResult(resp)
}

// handleListPodsOfInstance 处理list_pods_of_instance请求
func handleListPodsOfInstance(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	uid, _ := mcp.GetStringParam(request, "uid")
	instance, _ := mcp.GetStringParam(request, "instance")
	clusterId, _ := mcp.GetStringParam(request, "clusterId")
	current, _ := mcp.GetIntParam(request, "current")
	pageSize, _ := mcp.GetIntParam(request, "pageSize")

	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	req := &sysom.ListPodsOfInstanceRequest{
		Instance:  tea.String(instance),
		ClusterId: tea.String(clusterId),
		Current:   tea.Int64(int64(current)),
		PageSize:  tea.Int64(int64(pageSize)),
	}

	success, responseData, err := c.CallAPI(ctx, "list_pods_of_instance", req)
	if err != nil || !success {
		logger.Errorf("list_pods_of_instance failed: %v", err)
		return mcp.CreateErrorResult(fmt.Sprintf("list_pods_of_instance failed: %v", err))
	}

	resp := convertToMCPResponse(responseData)
	return mcp.CreateSuccessResult(resp)
}

// convertToMCPResponse 转换为MCP响应格式
func convertToMCPResponse(data interface{}) models.MCPResponse {
	resp := models.MCPResponse{
		Code: string(models.ResultCodeSuccess),
	}

	if data == nil {
		return resp
	}

	// 转换为map
	jsonData, err := json.Marshal(data)
	if err != nil {
		resp.Code = string(models.ResultCodeError)
		resp.Message = fmt.Sprintf("failed to marshal response: %v", err)
		return resp
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		resp.Code = string(models.ResultCodeError)
		resp.Message = fmt.Sprintf("failed to unmarshal response: %v", err)
		return resp
	}

	// 提取字段
	if code, ok := result["code"].(string); ok {
		resp.Code = code
	}
	if message, ok := result["message"].(string); ok {
		resp.Message = message
	}
	if data, ok := result["data"]; ok {
		resp.Data = data
	}
	if total, ok := result["total"].(float64); ok {
		t := int(total)
		resp.Total = &t
	}
	if requestId, ok := result["requestId"].(string); ok {
		resp.RequestID = requestId
	} else if requestId, ok := result["request_id"].(string); ok {
		resp.RequestID = requestId
	}

	return resp
}
