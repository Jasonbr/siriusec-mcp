package unit

import (
	"reflect"
	"siriusec-mcp/internal/registry"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock types for testing
type MockRequest struct {
	Name string
}

type MockResponse struct {
	Result string
}

func TestGetRegistry(t *testing.T) {
	reg1 := registry.GetRegistry()
	reg2 := registry.GetRegistry()

	// 应该返回同一个实例（单例模式）
	assert.Equal(t, reg1, reg2)
	assert.NotNil(t, reg1)
}

func TestRegisterSDK(t *testing.T) {
	reg := registry.GetRegistry()

	// 注册一个测试路由
	reg.RegisterSDK(
		"test_api",
		reflect.TypeOf(&MockRequest{}),
		reflect.TypeOf(&MockResponse{}),
		func(c interface{}, req interface{}) (interface{}, error) {
			return &MockResponse{Result: "success"}, nil
		},
	)

	// 验证路由已注册
	route := reg.GetRoute("test_api")
	assert.NotNil(t, route)
	assert.Equal(t, "test_api", route.APIName)
	assert.NotNil(t, route.SDKRoute)

	// 验证请求模型
	reqModel := reg.GetRequestModel("test_api")
	assert.NotNil(t, reqModel)
	assert.Equal(t, "MockRequest", reqModel.Elem().Name())

	// 验证响应模型
	respModel := reg.GetResponseModel("test_api")
	assert.NotNil(t, respModel)
	assert.Equal(t, "MockResponse", respModel.Elem().Name())
}

func TestGetRouteNotFound(t *testing.T) {
	reg := registry.GetRegistry()
	route := reg.GetRoute("non_existent_api")
	assert.Nil(t, route)
}
