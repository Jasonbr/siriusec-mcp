/*
诊断Helper实现

负责诊断相关的MCP工具逻辑，包括参数转换、调用诊断接口、轮询查询结果
*/
package diagnosis

import (
	"context"
	"encoding/json"
	"fmt"
	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/logger"
	"time"
)

// Helper 诊断Helper
type Helper struct {
	client       client.OpenAPIClient
	timeout      time.Duration
	pollInterval time.Duration
}

// NewHelper 创建诊断Helper
func NewHelper(c client.OpenAPIClient, timeout time.Duration, pollInterval time.Duration) *Helper {
	if timeout == 0 {
		timeout = 150 * time.Second
	}
	if pollInterval == 0 {
		pollInterval = 1 * time.Second
	}
	return &Helper{
		client:       c,
		timeout:      timeout,
		pollInterval: pollInterval,
	}
}

// Execute 执行诊断流程
func (h *Helper) Execute(ctx context.Context, req *Request) *Response {
	// 1. 准备参数并发起诊断
	params := make(map[string]interface{})
	for k, v := range req.Params {
		params[k] = v
	}
	if _, ok := params["source"]; !ok {
		params["source"] = "mcp"
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return NewErrorResponse(ResultCodeTaskCreateFailed, fmt.Sprintf("failed to marshal params: %v", err))
	}

	// 调用invoke_diagnosis接口
	invokeReq := map[string]interface{}{
		"service_name": req.ServiceName,
		"channel":      req.Channel,
		"params":       string(paramsJSON),
	}

	success, responseData, err := h.client.CallAPI(ctx, "invoke_diagnosis", invokeReq)
	if err != nil || !success {
		return NewErrorResponse(ResultCodeTaskCreateFailed, fmt.Sprintf("failed to invoke diagnosis: %v", err))
	}

	// 解析响应获取task_id
	respMap, ok := responseData.(map[string]interface{})
	if !ok {
		return NewErrorResponse(ResultCodeTaskCreateFailed, "invalid response format")
	}

	data, ok := respMap["data"].(map[string]interface{})
	if !ok {
		return NewErrorResponse(ResultCodeTaskCreateFailed, "invalid response data format")
	}

	taskID, ok := data["task_id"].(string)
	if !ok {
		return NewErrorResponse(ResultCodeTaskCreateFailed, "task_id not found in response")
	}

	logger.Infof("Diagnosis task created: %s", taskID)

	// 2. 轮询获取结果
	return h.waitForResult(ctx, taskID)
}

// waitForResult 轮询等待诊断结果
func (h *Helper) waitForResult(ctx context.Context, taskID string) *Response {
	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	ticker := time.NewTicker(h.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return NewErrorResponse(ResultCodeTaskTimeout, fmt.Sprintf("diagnosis timeout after %v, task_id: %s", h.timeout, taskID))

		case <-ticker.C:
			// 调用get_diagnosis_result接口
			getResultReq := map[string]interface{}{
				"task_id": taskID,
			}

			success, responseData, err := h.client.CallAPI(ctx, "get_diagnosis_result", getResultReq)
			if err != nil || !success {
				logger.Errorf("Failed to get diagnosis result: %v", err)
				continue
			}

			respMap, ok := responseData.(map[string]interface{})
			if !ok {
				continue
			}

			data, ok := respMap["data"].(map[string]interface{})
			if !ok {
				continue
			}

			taskStatus, ok := data["status"].(string)
			if !ok {
				continue
			}

			switch taskStatus {
			case "Fail":
				errMsg, _ := data["err_msg"].(string)
				return NewErrorResponse(ResultCodeTaskExecuteFailed, errMsg)

			case "Success":
				result := data["result"]
				var resultMap map[string]interface{}

				switch r := result.(type) {
				case string:
					if err := json.Unmarshal([]byte(r), &resultMap); err != nil {
						return NewErrorResponse(ResultCodeResultParseFailed, fmt.Sprintf("failed to parse result: %v", err))
					}
				case map[string]interface{}:
					resultMap = r
				default:
					resultMap = map[string]interface{}{"raw": result}
				}

				return NewSuccessResponse(taskID, resultMap)

			default:
				// 继续轮询
				logger.Debugf("Task %s status: %s, continuing to poll...", taskID, taskStatus)
			}
		}
	}
}
