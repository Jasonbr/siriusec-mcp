/*
阿里云SDK客户端实现

基于阿里云OpenAPI SDK的客户端实现
*/
package client

import (
	"context"
	"fmt"
	"reflect"
	"siriusec-mcp/internal/config"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/registry"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sysom "github.com/alibabacloud-go/sysom-20231230/client"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	// Endpoint 阿里云SysOM服务端点
	Endpoint = "sysom.cn-hangzhou.aliyuncs.com"
	// ConnectTimeout 连接超时时间(毫秒)
	ConnectTimeout = 2000
)

// AlibabaCloudSDKClient 基于阿里云OpenAPI SDK的客户端实现
type AlibabaCloudSDKClient struct {
	mode            string
	accessKeyID     string
	accessKeySecret string
	securityToken   string
	regionID        string
	client          *sysom.Client
}

// NewAlibabaCloudSDKClient 创建阿里云SDK客户端
func NewAlibabaCloudSDKClient(
	mode string,
	accessKeyID string,
	accessKeySecret string,
	securityToken string,
	regionID string,
) (*AlibabaCloudSDKClient, error) {
	if regionID == "" {
		regionID = "cn-hangzhou"
	}

	return &AlibabaCloudSDKClient{
		mode:            mode,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		securityToken:   securityToken,
		regionID:        regionID,
	}, nil
}

// GetSDKClient 获取或创建SDK客户端（实现OpenAPIClient接口）
func (c *AlibabaCloudSDKClient) GetSDKClient() *sysom.Client {
	client, err := c.getClient()
	if err != nil {
		logger.Errorf("Failed to get SDK client: %v", err)
		return nil
	}
	return client
}

// getClient 获取或创建SDK客户端（懒加载）
func (c *AlibabaCloudSDKClient) getClient() (*sysom.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(c.accessKeyID),
		AccessKeySecret: tea.String(c.accessKeySecret),
		Endpoint:        tea.String(Endpoint),
	}

	if c.securityToken != "" {
		config.SecurityToken = tea.String(c.securityToken)
	}

	client, err := sysom.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SysOM client: %w", err)
	}

	c.client = client
	return client, nil
}

// CallAPI 调用OpenAPI接口
func (c *AlibabaCloudSDKClient) CallAPI(ctx context.Context, apiName string, request interface{}) (bool, interface{}, error) {
	// 获取SDK路由
	route := registry.GetRegistry().GetSDKRoute(apiName)
	if route == nil {
		return false, nil, fmt.Errorf("API %s not registered", apiName)
	}

	// 获取客户端
	client, err := c.getClient()
	if err != nil {
		return false, nil, err
	}

	// 调用SDK方法
	response, err := route.ClientMethod(client, request)
	if err != nil {
		logger.Errorf("API call failed: %v", err)
		return false, nil, fmt.Errorf("API call failed: %w", err)
	}

	// 解析响应
	return c.parseResponse(response)
}

// parseResponse 解析SDK响应
func (c *AlibabaCloudSDKClient) parseResponse(response interface{}) (bool, interface{}, error) {
	if response == nil {
		return false, nil, fmt.Errorf("response is nil")
	}

	// 使用反射获取响应的 Body 字段
	v := reflect.ValueOf(response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 获取 Body 字段
	bodyField := v.FieldByName("Body")
	if !bodyField.IsValid() {
		return false, nil, fmt.Errorf("response body not found")
	}

	body := bodyField.Interface()

	// 检查状态码
	statusCodeField := v.FieldByName("StatusCode")
	if statusCodeField.IsValid() {
		statusCode := int(statusCodeField.Int())
		if statusCode != 200 {
			return false, body, fmt.Errorf("API returned status code %d", statusCode)
		}
	}

	return true, body, nil
}

// ClientFactory 客户端工厂
type ClientFactory struct{}

// NewClientFactory 创建客户端工厂
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// CreateClient 创建OpenAPI客户端
func (f *ClientFactory) CreateClient(uid string) (OpenAPIClient, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	mode := cfg.OpenAPI.Type
	accessKeyID := cfg.OpenAPI.AccessKeyID
	accessKeySecret := cfg.OpenAPI.AccessKeySecret
	securityToken := cfg.OpenAPI.SecurityToken
	regionID := cfg.OpenAPI.RegionID

	logger.Infof("Creating OpenAPI client, mode: %s, region: %s", mode, regionID)

	// 处理 RAM Role ARN 模式
	if mode == "ram_role_arn" {
		// 这里简化处理，实际应该使用STS AssumeRole
		logger.Warn("RAM Role ARN mode not fully implemented, using access_key mode")
		mode = "access_key"
	}

	return NewAlibabaCloudSDKClient(
		mode,
		accessKeyID,
		accessKeySecret,
		securityToken,
		regionID,
	)
}

// CreateClientWithOptions 使用选项创建客户端
func (f *ClientFactory) CreateClientWithOptions(
	mode string,
	accessKeyID string,
	accessKeySecret string,
	securityToken string,
	regionID string,
) (OpenAPIClient, error) {
	return NewAlibabaCloudSDKClient(mode, accessKeyID, accessKeySecret, securityToken, regionID)
}
