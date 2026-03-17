/*
崩溃诊断代理工具

提供vmcore诊断任务管理功能
*/
package crashagent

import (
	"context"
	"encoding/json"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"

	sysom "github.com/alibabacloud-go/sysom-20231230/client"
	"github.com/alibabacloud-go/tea/tea"
	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// TaskBaseInfo 任务基础信息
type TaskBaseInfo struct {
	TaskId     string `json:"taskId"`
	TaskType   string `json:"taskType"`
	TaskStatus string `json:"taskStatus"`
	CreatedAt  string `json:"createdAt"`
	ErrorMsg   string `json:"errorMsg"`
}

// TaskInfo 任务详情
type TaskInfo struct {
	TaskBaseInfo
	DiagnoseResult string                 `json:"diagnoseResult"`
	Urls           map[string]interface{} `json:"urls"`
}

// TaskId 任务ID
type TaskId struct {
	TaskId string `json:"taskId"`
}

// BaseResponse 基础响应
type BaseResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateTaskResponse 创建任务响应
type CreateTaskResponse struct {
	BaseResponse
	Data *TaskId `json:"data"`
}

// GetTaskResponse 获取任务详情响应
type GetTaskResponse struct {
	BaseResponse
	Data *TaskInfo `json:"data"`
}

// ListTasksResponse 任务列表响应
type ListTasksResponse struct {
	BaseResponse
	Data []TaskBaseInfo `json:"data"`
}

// RegisterTools 注册崩溃诊断代理工具
func RegisterTools(s *mcp.Server) {
	// list_vmcores (对应Python的list_history_tasks)
	s.RegisterTool("list_vmcores",
		mcpp.NewTool("list_vmcores",
			mcpp.WithDescription("查询历史创建的宕机诊断任务记录，返回指定天数内的任务列表"),
			mcpp.WithNumber("days", mcpp.Required(), mcpp.Description("查询几天前的历史任务记录，取值范围1-30天")),
		),
		handleListVmcores,
		"sysom_crash_agent",
	)

	// get_vmcore_detail (对应Python的query_diagnosis_task)
	s.RegisterTool("get_vmcore_detail",
		mcpp.NewTool("get_vmcore_detail",
			mcpp.WithDescription("查询诊断任务的结果，根据任务ID获取诊断任务的执行状态和结果"),
			mcpp.WithString("task_id", mcpp.Required(), mcpp.Description("诊断任务ID")),
		),
		handleGetVmcoreDetail,
		"sysom_crash_agent",
	)

	// analyze_vmcore (对应Python的create_vmcore_diagnosis_task)
	s.RegisterTool("analyze_vmcore",
		mcpp.NewTool("analyze_vmcore",
			mcpp.WithDescription("创建基于VMCORE文件的内核宕机诊断任务"),
			mcpp.WithString("vmcore_url", mcpp.Required(), mcpp.Description("vmcore文件下载链接")),
			mcpp.WithString("debuginfo_url", mcpp.Description("debuginfo文件下载链接（可选）")),
			mcpp.WithString("debuginfo_common_url", mcpp.Description("debuginfo-common文件下载链接（可选）")),
		),
		handleAnalyzeVmcore,
		"sysom_crash_agent",
	)

	// delete_vmcore (对应Python的create_dmesg_diagnosis_task - 简化处理)
	s.RegisterTool("delete_vmcore",
		mcpp.NewTool("delete_vmcore",
			mcpp.WithDescription("创建基于dmesg日志的系统诊断任务，分析系统宕机原因"),
			mcpp.WithString("dmesg_url", mcpp.Required(), mcpp.Description("dmesg日志文件下载链接")),
		),
		handleDeleteVmcore,
		"sysom_crash_agent",
	)
}

// handleListVmcores 处理查询历史任务请求
func handleListVmcores(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	daysFloat, ok := mcp.GetFloat64Param(request, "days")
	if !ok {
		return mcp.CreateErrorResult("days参数是必需的")
	}
	days := int32(daysFloat)

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient("")
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 调用API
	req := &sysom.ListVmcoreDiagnosisTaskRequest{
		Days: tea.Int64(int64(days)),
	}

	resp, err := c.GetSDKClient().ListVmcoreDiagnosisTask(req)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("查询历史任务失败: %v", err))
	}

	// 转换响应
	result := &ListTasksResponse{
		BaseResponse: BaseResponse{
			Code:    tea.StringValue(resp.Body.Code),
			Message: tea.StringValue(resp.Body.Message),
		},
	}

	if resp.Body.Data != nil {
		for _, item := range resp.Body.Data {
			result.Data = append(result.Data, TaskBaseInfo{
				TaskId:     tea.StringValue(item.TaskId),
				TaskType:   tea.StringValue(item.TaskType),
				TaskStatus: tea.StringValue(item.TaskStatus),
				CreatedAt:  tea.StringValue(item.CreatedAt),
				ErrorMsg:   tea.StringValue(item.ErrorMsg),
			})
		}
	}

	data, _ := json.Marshal(result)
	return mcp.CreateSuccessResult(string(data))
}

// handleGetVmcoreDetail 处理查询任务详情请求
func handleGetVmcoreDetail(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	taskID, ok := mcp.GetStringParam(request, "task_id")
	if !ok {
		return mcp.CreateErrorResult("task_id参数是必需的")
	}

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient("")
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 调用API
	req := &sysom.GetVmcoreDiagnosisTaskRequest{
		TaskId: tea.String(taskID),
	}

	resp, err := c.GetSDKClient().GetVmcoreDiagnosisTask(req)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("查询任务详情失败: %v", err))
	}

	// 转换响应
	result := &GetTaskResponse{
		BaseResponse: BaseResponse{
			Code:    tea.StringValue(resp.Body.Code),
			Message: tea.StringValue(resp.Body.Message),
		},
	}

	if resp.Body.Data != nil {
		data := resp.Body.Data
		urls := make(map[string]interface{})
		if data.Urls != nil {
			urls["vmcoreUrl"] = tea.StringValue(data.Urls.VmcoreUrl)
			urls["debuginfoUrl"] = tea.StringValue(data.Urls.DebuginfoUrl)
			urls["debuginfoCommonUrl"] = tea.StringValue(data.Urls.DebuginfoCommonUrl)
			urls["dmesgUrl"] = tea.StringValue(data.Urls.DmesgUrl)
		}
		result.Data = &TaskInfo{
			TaskBaseInfo: TaskBaseInfo{
				TaskId:     tea.StringValue(data.TaskId),
				TaskType:   tea.StringValue(data.TaskType),
				TaskStatus: tea.StringValue(data.TaskStatus),
				CreatedAt:  tea.StringValue(data.CreatedAt),
				ErrorMsg:   tea.StringValue(data.ErrorMsg),
			},
			DiagnoseResult: tea.StringValue(data.DiagnoseResult),
			Urls:           urls,
		}
	}

	data, _ := json.Marshal(result)
	return mcp.CreateSuccessResult(string(data))
}

// handleAnalyzeVmcore 处理创建vmcore诊断任务请求
func handleAnalyzeVmcore(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	vmcoreURL, ok := mcp.GetStringParam(request, "vmcore_url")
	if !ok {
		return mcp.CreateErrorResult("vmcore_url参数是必需的")
	}

	debuginfoURL, _ := mcp.GetStringParam(request, "debuginfo_url")
	debuginfoCommonURL, _ := mcp.GetStringParam(request, "debuginfo_common_url")

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient("")
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 调用API
	req := &sysom.CreateVmcoreDiagnosisTaskRequest{
		TaskType:           tea.String("vmcore"),
		VmcoreUrl:          tea.String(vmcoreURL),
		DebuginfoUrl:       tea.String(debuginfoURL),
		DebuginfoCommonUrl: tea.String(debuginfoCommonURL),
	}

	resp, err := c.GetSDKClient().CreateVmcoreDiagnosisTask(req)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("创建诊断任务失败: %v", err))
	}

	// 转换响应
	result := &CreateTaskResponse{
		BaseResponse: BaseResponse{
			Code:    tea.StringValue(resp.Body.Code),
			Message: tea.StringValue(resp.Body.Message),
		},
	}

	if resp.Body.Data != nil {
		result.Data = &TaskId{
			TaskId: tea.StringValue(resp.Body.Data.TaskId),
		}
	}

	data, _ := json.Marshal(result)
	return mcp.CreateSuccessResult(string(data))
}

// handleDeleteVmcore 处理创建dmesg诊断任务请求
func handleDeleteVmcore(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	dmesgURL, ok := mcp.GetStringParam(request, "dmesg_url")
	if !ok {
		return mcp.CreateErrorResult("dmesg_url参数是必需的")
	}

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient("")
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 调用API
	req := &sysom.CreateVmcoreDiagnosisTaskRequest{
		TaskType: tea.String("dmesg"),
		DmesgUrl: tea.String(dmesgURL),
	}

	resp, err := c.GetSDKClient().CreateVmcoreDiagnosisTask(req)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("创建诊断任务失败: %v", err))
	}

	// 转换响应
	result := &CreateTaskResponse{
		BaseResponse: BaseResponse{
			Code:    tea.StringValue(resp.Body.Code),
			Message: tea.StringValue(resp.Body.Message),
		},
	}

	if resp.Body.Data != nil {
		result.Data = &TaskId{
			TaskId: tea.StringValue(resp.Body.Data.TaskId),
		}
	}

	data, _ := json.Marshal(result)
	return mcp.CreateSuccessResult(string(data))
}
