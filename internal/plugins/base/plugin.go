/*
监控插件框架

提供统一的插件接口和生命周期管理
*/
package base

import (
	"context"
	"time"
)

// PluginStatus 插件状态
type PluginStatus string

const (
	StatusRunning  PluginStatus = "running"
	StatusStopped  PluginStatus = "stopped"
	StatusError    PluginStatus = "error"
	StatusStarting PluginStatus = "starting"
	StatusStopping PluginStatus = "stopping"
)

// Event 监控事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`     // 事件类型：cpu_high, mem_high, disk_full...
	Severity  string                 `json:"severity"` // 严重程度：critical, warning, info
	Timestamp time.Time              `json:"timestamp"`
	Target    string                 `json:"target"`    // 目标：instance_id, hostname...
	Metric    string                 `json:"metric"`    // 指标名
	Value     interface{}            `json:"value"`     // 当前值
	Threshold interface{}            `json:"threshold"` // 阈值
	Message   string                 `json:"message"`   // 人类可读的描述
	Labels    map[string]string      `json:"labels"`    // 标签
	RawData   map[string]interface{} `json:"raw_data"`  // 原始数据
}

// PluginConfig 插件配置
type PluginConfig struct {
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Interval time.Duration          `json:"interval"` // 检查间隔
	Params   map[string]interface{} `json:"params"`   // 插件参数
}

// CheckPlugin 检查插件接口（主动监控）
type CheckPlugin interface {
	// Name 插件名称
	Name() string

	// Description 插件描述
	Description() string

	// Init 初始化插件
	Init(config PluginConfig) error

	// Start 启动监控循环
	Start(ctx context.Context) error

	// Stop 停止监控
	Stop() error

	// Status 获取插件状态
	Status() PluginStatus

	// Check 执行一次检查
	Check(ctx context.Context) (*Event, error)
}

// DiagnosticTool 诊断工具接口（被动调用）
type DiagnosticTool interface {
	// Name 工具名称
	Name() string

	// Description 工具描述
	Description() string

	// Category 工具分类
	Category() string

	// Execute 执行诊断
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Notifier 通知器接口
type Notifier interface {
	// Name 通知器名称
	Name() string

	// Send 发送通知
	Send(event *Event) error

	// Close 关闭连接
	Close() error
}
