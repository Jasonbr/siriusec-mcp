/*
初始化服务工具

提供check_sysom_initialed和initial_sysom功能
*/
package initial

import (
	"context"
	"encoding/json"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/mcp"
	"sync"

	sysom "github.com/alibabacloud-go/sysom-20231230/client"
	"github.com/alibabacloud-go/tea/tea"
	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// 全局缓存
var (
	sysomInitialedCache = make(map[string]bool)
	cacheMutex          sync.RWMutex
)

// ResultCode 结果状态码
type ResultCode string

const (
	ResultCodeSuccess ResultCode = "Success"
	ResultCodeError   ResultCode = "Error"
)

// InitialResponse 初始化响应
type InitialResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// getSysomInitialedStatus 从缓存获取开通状态
func getSysomInitialedStatus(uid string) (bool, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	status, exists := sysomInitialedCache[uid]
	return status, exists
}

// setSysomInitialedStatus 设置开通状态到缓存
func setSysomInitialedStatus(uid string, isInitialed bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	sysomInitialedCache[uid] = isInitialed
}

// RegisterTools 注册初始化服务工具
func RegisterTools(s *mcp.Server) {
	// check_sysom_initialed
	s.RegisterTool("check_sysom_initialed",
		mcpp.NewTool("check_sysom_initialed",
			mcpp.WithDescription("检查sysom服务是否已开通"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
		),
		handleCheckSysomInitialed,
		"sysom_initial",
	)

	// initial_sysom
	s.RegisterTool("initial_sysom",
		mcpp.NewTool("initial_sysom",
			mcpp.WithDescription("帮助用户开通sysom服务"),
			mcpp.WithString("uid", mcpp.Required(), mcpp.Description("用户ID")),
		),
		handleInitialSysom,
		"sysom_initial",
	)
}

// handleCheckSysomInitialed 处理检查开通状态请求
func handleCheckSysomInitialed(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	uid, ok := mcp.GetStringParam(request, "uid")
	if !ok {
		return mcp.CreateErrorResult("uid参数是必需的")
	}

	// 先检查缓存
	cachedStatus, exists := getSysomInitialedStatus(uid)
	if exists && cachedStatus {
		result := &InitialResponse{
			Code:    string(ResultCodeSuccess),
			Message: "用户已开通sysom服务（来自缓存）",
			Data:    map[string]interface{}{"initialed": true, "from_cache": true},
		}
		data, _ := json.Marshal(result)
		return mcp.CreateSuccessResult(string(data))
	}

	// 缓存中没有或显示未开通，调用API检查
	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 调用API
	req := &sysom.InitialSysomRequest{
		CheckOnly: tea.Bool(true),
		Source:    tea.String("mcp"),
	}

	resp, err := c.GetSDKClient().InitialSysom(req)
	if err != nil {
		setSysomInitialedStatus(uid, false)
		result := &InitialResponse{
			Code:    string(ResultCodeError),
			Message: fmt.Sprintf("检查开通状态失败: %v", err),
			Data:    nil,
		}
		data, _ := json.Marshal(result)
		return mcp.CreateSuccessResult(string(data))
	}

	// 更新缓存
	isSuccess := tea.StringValue(resp.Body.Code) == "Success"
	setSysomInitialedStatus(uid, isSuccess)

	result := &InitialResponse{
		Code:    tea.StringValue(resp.Body.Code),
		Message: tea.StringValue(resp.Body.Message),
		Data:    resp.Body.Data,
	}
	data, _ := json.Marshal(result)
	return mcp.CreateSuccessResult(string(data))
}

// handleInitialSysom 处理开通服务请求
func handleInitialSysom(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	uid, ok := mcp.GetStringParam(request, "uid")
	if !ok {
		return mcp.CreateErrorResult("uid参数是必需的")
	}

	// 创建客户端
	c, err := client.GlobalFactory.CreateClient(uid)
	if err != nil {
		return mcp.CreateErrorResult(fmt.Sprintf("failed to create client: %v", err))
	}

	// 调用API
	req := &sysom.InitialSysomRequest{
		CheckOnly: tea.Bool(false),
		Source:    tea.String("mcp"),
	}

	resp, err := c.GetSDKClient().InitialSysom(req)
	if err != nil {
		result := &InitialResponse{
			Code:    string(ResultCodeError),
			Message: fmt.Sprintf("开通sysom失败: %v", err),
			Data:    nil,
		}
		data, _ := json.Marshal(result)
		return mcp.CreateSuccessResult(string(data))
	}

	// 更新缓存
	isSuccess := tea.StringValue(resp.Body.Code) == "Success"
	setSysomInitialedStatus(uid, isSuccess)

	result := &InitialResponse{
		Code:    tea.StringValue(resp.Body.Code),
		Message: tea.StringValue(resp.Body.Message),
		Data:    resp.Body.Data,
	}
	data, _ := json.Marshal(result)
	return mcp.CreateSuccessResult(string(data))
}
