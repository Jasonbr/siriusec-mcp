/*
API路由注册表

统一管理OpenAPI的URL到模型和方法的映射关系
*/
package registry

import (
	"reflect"
	"sync"
)

// SDKRoute SDK调用方式的路由信息
type SDKRoute struct {
	RequestModel  reflect.Type
	ResponseModel reflect.Type
	ClientMethod  func(client interface{}, request interface{}) (interface{}, error)
}

// APIRoute API路由信息
type APIRoute struct {
	APIName   string
	SDKRoute  *SDKRoute
}

// Registry API路由注册表（单例模式）
type Registry struct {
	routes map[string]*APIRoute
	mu     sync.RWMutex
}

var (
	instance *Registry
	once     sync.Once
)

// GetRegistry 获取注册表单例
func GetRegistry() *Registry {
	once.Do(func() {
		instance = &Registry{
			routes: make(map[string]*APIRoute),
		}
	})
	return instance
}

// RegisterSDK 注册SDK调用方式的路由
func (r *Registry) RegisterSDK(
	apiName string,
	requestModel reflect.Type,
	responseModel reflect.Type,
	clientMethod func(client interface{}, request interface{}) (interface{}, error),
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sdkRoute := &SDKRoute{
		RequestModel:  requestModel,
		ResponseModel: responseModel,
		ClientMethod:  clientMethod,
	}

	if route, exists := r.routes[apiName]; exists {
		route.SDKRoute = sdkRoute
	} else {
		r.routes[apiName] = &APIRoute{
			APIName:  apiName,
			SDKRoute: sdkRoute,
		}
	}
}

// GetRoute 根据接口名称获取路由信息
func (r *Registry) GetRoute(apiName string) *APIRoute {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.routes[apiName]
}

// GetSDKRoute 获取SDK路由信息
func (r *Registry) GetSDKRoute(apiName string) *SDKRoute {
	route := r.GetRoute(apiName)
	if route != nil {
		return route.SDKRoute
	}
	return nil
}

// GetRequestModel 获取请求模型
func (r *Registry) GetRequestModel(apiName string) reflect.Type {
	sdkRoute := r.GetSDKRoute(apiName)
	if sdkRoute != nil {
		return sdkRoute.RequestModel
	}
	return nil
}

// GetResponseModel 获取响应模型
func (r *Registry) GetResponseModel(apiName string) reflect.Type {
	sdkRoute := r.GetSDKRoute(apiName)
	if sdkRoute != nil {
		return sdkRoute.ResponseModel
	}
	return nil
}
