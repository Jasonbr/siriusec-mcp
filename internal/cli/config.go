// Package cli 提供命令行工具功能
package cli

import (
	"fmt"
	"os"
	"siriusec-mcp/internal/config"
)

// ValidateConfig 验证配置
func ValidateConfig(configDir string) error {
	fmt.Println("Validating configuration...")
	fmt.Println()

	// 检查配置文件目录
	if configDir == "" {
		configDir = "."
	}

	info, err := os.Stat(configDir)
	if err != nil {
		return fmt.Errorf("config directory not found: %s", configDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("config path is not a directory: %s", configDir)
	}

	fmt.Printf("✓ Config directory: %s\n", configDir)

	// 检查 .env 文件
	envFile := configDir + "/.env"
	if _, err := os.Stat(envFile); err == nil {
		fmt.Printf("✓ Environment file found: %s\n", envFile)
	} else {
		fmt.Printf("⚠ Environment file not found: %s\n", envFile)
	}

	// 加载配置
	cfg := config.LoadConfig()

	// 验证 OpenAPI 配置
	fmt.Println()
	fmt.Println("OpenAPI Configuration:")
	if cfg.OpenAPI.AccessKeyID == "" {
		fmt.Printf("  ✗ ACCESS_KEY_ID is not set\n")
	} else {
		masked := maskString(cfg.OpenAPI.AccessKeyID, 4)
		fmt.Printf("  ✓ Access Key ID: %s\n", masked)
	}

	if cfg.OpenAPI.AccessKeySecret == "" {
		fmt.Printf("  ✗ ACCESS_KEY_SECRET is not set\n")
	} else {
		masked := maskString(cfg.OpenAPI.AccessKeySecret, 4)
		fmt.Printf("  ✓ Access Key Secret: %s\n", masked)
	}

	fmt.Printf("  ✓ Region: %s\n", cfg.OpenAPI.RegionID)
	fmt.Printf("  ✓ Type: %s\n", cfg.OpenAPI.Type)

	// 验证日志配置
	fmt.Println()
	fmt.Println("Log Configuration:")
	fmt.Printf("  ✓ Level: %s\n", cfg.Log.Level)

	// 验证 LLM 配置（可选）
	fmt.Println()
	fmt.Println("LLM Configuration (optional):")
	if cfg.LLM.DashScopeAPIKey == "" {
		fmt.Printf("  ⚠ DASHSCOPE_API_KEY is not set (optional)\n")
	} else {
		masked := maskString(cfg.LLM.DashScopeAPIKey, 4)
		fmt.Printf("  ✓ DashScope API Key: %s\n", masked)
	}

	fmt.Println()
	fmt.Println("Configuration validation completed.")

	return nil
}

// maskString 遮罩字符串，只显示前后 n 个字符
func maskString(s string, n int) string {
	if len(s) <= n*2 {
		return "***"
	}
	return s[:n] + "***" + s[len(s)-n:]
}
