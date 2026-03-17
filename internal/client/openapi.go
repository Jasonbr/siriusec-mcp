/*
OpenAPI客户端接口定义

提供统一的OpenAPI调用接口
*/
package client

import (
	"context"

	sysom "github.com/alibabacloud-go/sysom-20231230/client"
)

// Response 通用响应结构
type Response struct {
	Success bool
	Data    interface{}
	Error   string
}

// OpenAPIClient OpenAPI客户端接口
type OpenAPIClient interface {
	// CallAPI 调用OpenAPI接口
	// 返回: (是否成功, 响应数据, 错误信息)
	CallAPI(ctx context.Context, apiName string, request interface{}) (bool, interface{}, error)

	// GetSDKClient 获取底层SDK客户端
	GetSDKClient() *sysom.Client
}
