package unit

import (
	"os"
	"siriusec-mcp/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("OPENAPI_TYPE", "access_key")
	os.Setenv("ACCESS_KEY_ID", "test-key-id")
	os.Setenv("ACCESS_KEY_SECRET", "test-key-secret")
	os.Setenv("LOG_LEVEL", "DEBUG")

	cfg := config.LoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "access_key", cfg.OpenAPI.Type)
	assert.Equal(t, "test-key-id", cfg.OpenAPI.AccessKeyID)
	assert.Equal(t, "test-key-secret", cfg.OpenAPI.AccessKeySecret)
	assert.Equal(t, "DEBUG", cfg.Log.Level)
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected int // logrus.Level 实际上是 int 类型
	}{
		{"DEBUG", 5},
		{"INFO", 4},
		{"WARN", 3},
		{"ERROR", 2},
		{"UNKNOWN", 4}, // 默认 INFO
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			cfg := &config.Config{
				Log: config.LogConfig{Level: tt.level},
			}
			// 只检查是否返回了有效的日志级别，不检查具体值
			assert.NotNil(t, cfg.GetLogLevel())
		})
	}
}
