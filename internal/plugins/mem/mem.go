/*
内存监控插件

监控内存使用率、Swap 使用情况、OOM 检测等
*/
package mem

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"siriusec-mcp/internal/plugins/base"
)

// MemPlugin 内存监控插件
type MemPlugin struct {
	config        base.PluginConfig
	status        base.PluginStatus
	memThreshold  float64 // 内存使用率阈值（百分比）
	swapThreshold float64 // Swap 使用率阈值（百分比）
}

// Name 插件名称
func (p *MemPlugin) Name() string {
	return "mem"
}

// Description 插件描述
func (p *MemPlugin) Description() string {
	return "内存使用率、Swap 监控"
}

// Init 初始化插件
func (p *MemPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取阈值，默认内存 85%，Swap 80%
	if threshold, ok := config.Params["mem_threshold"].(float64); ok {
		p.memThreshold = threshold
	} else {
		p.memThreshold = 85.0
	}

	if threshold, ok := config.Params["swap_threshold"].(float64); ok {
		p.swapThreshold = threshold
	} else {
		p.swapThreshold = 80.0
	}

	return nil
}

// Start 启动监控循环
func (p *MemPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("mem check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[MEM Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *MemPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *MemPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *MemPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 获取内存信息
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// 获取 Swap 信息
	swapInfo, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap info: %w", err)
	}

	// 检查内存使用率
	if memInfo.UsedPercent > p.memThreshold {
		return p.createMemHighEvent(memInfo), nil
	}

	// 检查 Swap 使用率
	if swapInfo.UsedPercent > p.swapThreshold {
		return p.createSwapHighEvent(swapInfo), nil
	}

	return nil, nil
}

// createMemHighEvent 创建内存过高事件
func (p *MemPlugin) createMemHighEvent(memInfo *mem.VirtualMemoryStat) *base.Event {
	severity := "warning"
	if memInfo.UsedPercent >= 95 {
		severity = "critical"
	}

	return &base.Event{
		ID:        fmt.Sprintf("mem_high_%d", time.Now().Unix()),
		Type:      "mem_high",
		Severity:  severity,
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "memory_usage",
		Value:     memInfo.UsedPercent,
		Threshold: p.memThreshold,
		Message:   fmt.Sprintf("内存使用率过高：%.1f%% (阈值：%.1f%%)", memInfo.UsedPercent, p.memThreshold),
		Labels: map[string]string{
			"plugin":   "mem",
			"hostname": getHostname(),
			"severity": severity,
		},
		RawData: map[string]interface{}{
			"total":       memInfo.Total,
			"used":        memInfo.Used,
			"free":        memInfo.Free,
			"available":   memInfo.Available,
			"used_percent": memInfo.UsedPercent,
			"buffers":     memInfo.Buffers,
			"cached":      memInfo.Cached,
		},
	}
}

// createSwapHighEvent 创建 Swap 过高事件
func (p *MemPlugin) createSwapHighEvent(swapInfo *mem.SwapMemoryStat) *base.Event {
	severity := "warning"
	if swapInfo.UsedPercent >= 95 {
		severity = "critical"
	}

	return &base.Event{
		ID:        fmt.Sprintf("swap_high_%d", time.Now().Unix()),
		Type:      "swap_high",
		Severity:  severity,
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "swap_usage",
		Value:     swapInfo.UsedPercent,
		Threshold: p.swapThreshold,
		Message:   fmt.Sprintf("Swap 使用率过高：%.1f%% (阈值：%.1f%%)", swapInfo.UsedPercent, p.swapThreshold),
		Labels: map[string]string{
			"plugin":   "mem",
			"hostname": getHostname(),
			"severity": severity,
		},
		RawData: map[string]interface{}{
			"total":        swapInfo.Total,
			"used":         swapInfo.Used,
			"free":         swapInfo.Free,
			"used_percent": swapInfo.UsedPercent,
			"in":           swapInfo.Sin,
			"out":          swapInfo.Sout,
		},
	}
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
package mem
