/*
插件管理器

负责所有插件的注册、启动、停止和状态管理
*/
package base

import (
	"context"
	"fmt"
	"sync"
	"time"

	"siriusec-mcp/internal/logger"
)

// Manager 插件管理器
type Manager struct {
	mu        sync.RWMutex
	plugins   map[string]CheckPlugin
	tools     map[string]DiagnosticTool
	notifiers map[string]Notifier
	configs   map[string]PluginConfig
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewManager 创建新的插件管理器
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		plugins:   make(map[string]CheckPlugin),
		tools:     make(map[string]DiagnosticTool),
		notifiers: make(map[string]Notifier),
		configs:   make(map[string]PluginConfig),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// RegisterPlugin 注册监控插件
func (m *Manager) RegisterPlugin(plugin CheckPlugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	m.plugins[name] = plugin
	logger.Sugar.Infow("Plugin registered", "name", name)
	return nil
}

// RegisterTool 注册诊断工具
func (m *Manager) RegisterTool(tool DiagnosticTool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := tool.Name()
	if _, exists := m.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	m.tools[name] = tool
	logger.Sugar.Infow("Diagnostic tool registered", "name", name, "category", tool.Category())
	return nil
}

// RegisterNotifier 注册通知器
func (m *Manager) RegisterNotifier(notifier Notifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := notifier.Name()
	if _, exists := m.notifiers[name]; exists {
		return fmt.Errorf("notifier %s already registered", name)
	}

	m.notifiers[name] = notifier
	logger.Sugar.Infow("Notifier registered", "name", name)
	return nil
}

// StartPlugin 启动指定插件
func (m *Manager) StartPlugin(name string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 获取配置
	config, ok := m.configs[name]
	if !ok {
		// 使用默认配置
		config = PluginConfig{
			Name:     name,
			Enabled:  true,
			Interval: 30 * time.Second, // 默认 30 秒检查一次
		}
	}

	// 初始化插件
	if err := plugin.Init(config); err != nil {
		return fmt.Errorf("failed to init plugin %s: %w", name, err)
	}

	// 启动监控循环
	go func() {
		logger.Sugar.Infow("Starting plugin", "name", name)
		if err := plugin.Start(m.ctx); err != nil {
			logger.Sugar.Errorw("Plugin stopped with error", "name", name, "error", err)
		}
	}()

	return nil
}

// StopPlugin 停止指定插件
func (m *Manager) StopPlugin(name string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := plugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin %s: %w", name, err)
	}

	logger.Sugar.Infow("Plugin stopped", "name", name)
	return nil
}

// StartAllPlugins 启动所有已启用的插件
func (m *Manager) StartAllPlugins() error {
	for name := range m.plugins {
		if err := m.StartPlugin(name); err != nil {
			logger.Sugar.Errorw("Failed to start plugin", "name", name, "error", err)
		}
	}
	return nil
}

// StopAllPlugins 停止所有插件
func (m *Manager) StopAllPlugins() error {
	for name := range m.plugins {
		if err := m.StopPlugin(name); err != nil {
			logger.Sugar.Errorw("Failed to stop plugin", "name", name, "error", err)
		}
	}
	return nil
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (CheckPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []CheckPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]CheckPlugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// ListTools 列出所有诊断工具
func (m *Manager) ListTools() []DiagnosticTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]DiagnosticTool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetTool 获取诊断工具
func (m *Manager) GetTool(name string) (DiagnosticTool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, exists := m.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// Close 关闭管理器
func (m *Manager) Close() {
	m.cancel()
	m.StopAllPlugins()
}
