/*
客户端工厂

统一创建OpenAPI客户端实例
*/
package client

import (
	"fmt"
	"siriusec-mcp/internal/config"
	"siriusec-mcp/internal/logger"
)

// Factory 客户端工厂接口
type Factory interface {
	CreateClient(uid string) (OpenAPIClient, error)
}

// DefaultFactory 默认客户端工厂
type DefaultFactory struct{}

// NewFactory 创建默认工厂
func NewFactory() Factory {
	return &DefaultFactory{}
}

// CreateClient 创建OpenAPI客户端
func (f *DefaultFactory) CreateClient(uid string) (OpenAPIClient, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	mode := cfg.OpenAPI.Type
	accessKeyID := cfg.OpenAPI.AccessKeyID
	accessKeySecret := cfg.OpenAPI.AccessKeySecret
	securityToken := cfg.OpenAPI.SecurityToken
	regionID := cfg.OpenAPI.RegionID

	logger.Infof("Creating OpenAPI client, mode: %s, region: %s, uid: %s", mode, regionID, uid)

	// 处理 RAM Role ARN 模式
	if mode == "ram_role_arn" {
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

// GlobalFactory 全局工厂实例
var GlobalFactory Factory = NewFactory()

// SetGlobalFactory 设置全局工厂
func SetGlobalFactory(factory Factory) {
	GlobalFactory = factory
}
