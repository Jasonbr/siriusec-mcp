/*
配置管理模块

从 .env 文件读取配置，提供全局配置对象
*/
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config 全局配置
type Config struct {
	DeployMode string
	OpenAPI    OpenAPIConfig
	LLM        LLMConfig
	Log        LogConfig
}

// OpenAPIConfig OpenAPI配置
type OpenAPIConfig struct {
	Type            string
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
	RoleARN         string
	RegionID        string
}

// LLMConfig LLM配置
type LLMConfig struct {
	DashScopeAPIKey string
	LLMAK           string
}

// LogConfig 日志配置
type LogConfig struct {
	Level string
}

var (
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
)

func init() {
	GlobalConfig = LoadConfig()
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 尝试加载 .env 文件
	loadEnvFile()

	cfg := &Config{
		DeployMode: getEnv("DEPLOY_MODE", "alibabacloud_sdk"),
		OpenAPI: OpenAPIConfig{
			Type:            getEnv("OPENAPI_TYPE", getEnv("ACCESS_KEY_ID", "access_key")),
			AccessKeyID:     getEnv("OPENAPI_ACCESS_KEY_ID", getEnv("ACCESS_KEY_ID", "")),
			AccessKeySecret: getEnv("OPENAPI_ACCESS_KEY_SECRET", getEnv("ACCESS_KEY_SECRET", "")),
			SecurityToken:   getEnv("OPENAPI_SECURITY_TOKEN", getEnv("SECURITY_TOKEN", "")),
			RoleARN:         getEnv("OPENAPI_ROLE_ARN", getEnv("ROLE_ARN", "")),
			RegionID:        getEnv("REGION_ID", "cn-hangzhou"),
		},
		LLM: LLMConfig{
			DashScopeAPIKey: getEnv("DASHSCOPE_API_KEY", ""),
			LLMAK:           getEnv("sysom_service___llm___llm_ak", getEnv("DASHSCOPE_API_KEY", "")),
		},
		Log: LogConfig{
			Level: strings.ToUpper(getEnv("LOG_LEVEL", "INFO")),
		},
	}

	return cfg
}

// loadEnvFile 加载 .env 文件
func loadEnvFile() {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		// 静默处理，使用系统环境变量
		godotenv.Load()
		return
	}

	// 尝试在当前目录加载 .env
	envFile := filepath.Join(cwd, ".env")
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err == nil {
			return
		}
	}

	// 尝试在项目根目录加载 .env
	parentDir := filepath.Dir(cwd)
	envFile = filepath.Join(parentDir, ".env")
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err == nil {
			return
		}
	}

	// 尝试加载系统环境变量
	godotenv.Load()
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetLogLevel 获取日志级别字符串
func (c *Config) GetLogLevel() string {
	return c.Log.Level
}
