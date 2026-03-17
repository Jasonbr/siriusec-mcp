/*
诊断服务基类

提供诊断相关的通用结构和接口
*/
package diagnosis

import (
	"siriusec-mcp/pkg/models"
)

// ResultCode 诊断结果状态码
type ResultCode string

const (
	// ResultCodeSuccess 成功
	ResultCodeSuccess ResultCode = "Success"
	// ResultCodeTaskCreateFailed 任务创建失败
	ResultCodeTaskCreateFailed ResultCode = "TaskCreateFailed"
	// ResultCodeTaskExecuteFailed 任务执行失败
	ResultCodeTaskExecuteFailed ResultCode = "TaskExecuteFailed"
	// ResultCodeTaskTimeout 任务超时
	ResultCodeTaskTimeout ResultCode = "TaskTimeout"
	// ResultCodeResultParseFailed 结果解析失败
	ResultCodeResultParseFailed ResultCode = "ResultParseFailed"
	// ResultCodeGetResultFailed 获取结果失败
	ResultCodeGetResultFailed ResultCode = "GetResultFailed"
)

// Request 诊断请求
type Request struct {
	ServiceName string
	Channel     string
	Region      string
	Params      map[string]interface{}
}

// Response 诊断响应
type Response struct {
	Code    ResultCode
	Message string
	TaskID  string
	Result  map[string]interface{}
}

// ToMCPResponse 转换为MCP响应
func (r *Response) ToMCPResponse() models.DiagnosisResponse {
	return models.DiagnosisResponse{
		MCPResponse: models.MCPResponse{
			Code:    string(r.Code),
			Message: r.Message,
		},
		TaskID: r.TaskID,
		Result: r.Result,
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(taskID string, result map[string]interface{}) *Response {
	return &Response{
		Code:    ResultCodeSuccess,
		TaskID:  taskID,
		Result:  result,
		Message: "",
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code ResultCode, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		TaskID:  "",
		Result:  nil,
	}
}
