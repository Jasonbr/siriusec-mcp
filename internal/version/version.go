// Package version 提供版本信息
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version 版本号，通过 ldflags 注入
	Version = "dev"
	// GitCommit Git 提交哈希，通过 ldflags 注入
	GitCommit = "unknown"
	// BuildTime 构建时间，通过 ldflags 注入
	BuildTime = "unknown"
	// GoVersion Go 版本
	GoVersion = runtime.Version()
	// OS 操作系统
	OS = runtime.GOOS
	// Arch 架构
	Arch = runtime.GOARCH
)

// Info 返回版本信息字符串
func Info() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s, os: %s/%s)",
		Version, GitCommit, BuildTime, GoVersion, OS, Arch)
}

// Short 返回简短版本信息
func Short() string {
	return Version
}

// Full 返回完整版本信息
func Full() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_time": BuildTime,
		"go_version": GoVersion,
		"os":         OS,
		"arch":       Arch,
	}
}
