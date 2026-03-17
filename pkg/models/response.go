/*
响应模型定义

定义所有MCP工具使用的响应结构体
*/
package models

// ResultCode 结果状态码
type ResultCode string

const (
	// 通用状态码
	ResultCodeSuccess ResultCode = "Success"
	ResultCodeError   ResultCode = "Error"

	// 诊断状态码
	ResultCodeTaskCreateFailed  ResultCode = "TaskCreateFailed"
	ResultCodeTaskExecuteFailed ResultCode = "TaskExecuteFailed"
	ResultCodeTaskTimeout       ResultCode = "TaskTimeout"
	ResultCodeResultParseFailed ResultCode = "ResultParseFailed"
	ResultCodeGetResultFailed   ResultCode = "GetResultFailed"
)

// MCPResponse MCP响应基类
type MCPResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Total     *int        `json:"total,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// ListAllInstancesResponse 列出所有实例响应
type ListAllInstancesResponse struct {
	MCPResponse
	NextToken  string `json:"nextToken,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
}

// ListInstancesResponse 列出实例响应
type ListInstancesResponse struct {
	MCPResponse
}

// ListClustersResponse 列出集群响应
type ListClustersResponse struct {
	MCPResponse
}

// ListPodsOfInstanceResponse 列出实例下Pod响应
type ListPodsOfInstanceResponse struct {
	MCPResponse
}

// DiagnosisResponse 诊断响应
type DiagnosisResponse struct {
	MCPResponse
	TaskID string                 `json:"task_id,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
}

// CrashAgentResponse 崩溃诊断代理响应
type CrashAgentResponse struct {
	MCPResponse
}

// InitialSysomResponse 初始化Sysom响应
type InitialSysomResponse struct {
	MCPResponse
}

// CheckSysomInitialedResponse 检查Sysom初始化状态响应
type CheckSysomInitialedResponse struct {
	MCPResponse
	Initialed bool `json:"initialed,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, total ...int) MCPResponse {
	resp := MCPResponse{
		Code: string(ResultCodeSuccess),
		Data: data,
	}
	if len(total) > 0 {
		t := total[0]
		resp.Total = &t
	}
	return resp
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string) MCPResponse {
	return MCPResponse{
		Code:    string(ResultCodeError),
		Message: message,
	}
}

// NewDiagnosisResponse 创建诊断响应
func NewDiagnosisResponse(code ResultCode, message string, taskID string, result map[string]interface{}) DiagnosisResponse {
	return DiagnosisResponse{
		MCPResponse: MCPResponse{
			Code:    string(code),
			Message: message,
		},
		TaskID: taskID,
		Result: result,
	}
}

// IsSuccess 检查是否成功
func (r *MCPResponse) IsSuccess() bool {
	return r.Code == string(ResultCodeSuccess)
}

// IsPermissionError 检查是否是权限错误
func IsPermissionError(message string) bool {
	permissionKeywords := []string{
		"权限",
		"permission",
		"unauthorized",
		"access denied",
		"forbidden",
		"RAM",
		"授权",
		"authorize",
	}
	for _, keyword := range permissionKeywords {
		if contains(message, keyword) {
			return true
		}
	}
	return false
}

// EnhancePermissionErrorMessage 增强权限错误消息
func EnhancePermissionErrorMessage(message string) string {
	return message + "\n\n这可能是权限问题。请检查：\n" +
		"1. 您的阿里云账号是否具有访问 SysOM 服务的权限\n" +
		"2. 是否正确配置了 AccessKey ID 和 AccessKey Secret\n" +
		"3. 是否需要为 RAM 用户授予 AliyunSysOMFullAccess 权限\n" +
		"4. 如果使用 STS Token，请检查 Token 是否过期"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
