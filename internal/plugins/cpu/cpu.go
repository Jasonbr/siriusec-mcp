package cpu
/*
CPU 监控插件

监控 CPU 使用率、Load Average 等指标
*/
package cpu

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"siriusec-mcp/internal/plugins/base"
)

// CPUPlugin CPU 监控插件
type CPUPlugin struct {
	config    base.PluginConfig
	status    base.PluginStatus
	threshold float64 // CPU 使用率阈值（百分比）
}

// Name 插件名称
func (p *CPUPlugin) Name() string {
	return "cpu"
}

// Description 插件描述
func (p *CPUPlugin) Description() string {
	return "CPU 使用率、Load Average 监控"
}

// Init 初始化插件
func (p *CPUPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取阈值，默认 80%
	if threshold, ok := config.Params["threshold"].(float64); ok {
		p.threshold = threshold
	} else {
		p.threshold = 80.0
	}

	return nil
}

// Start 启动监控循环
func (p *CPUPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("cpu check failed: %w", err)
			}

			// 如果检测到异常，触发告警
			if event != nil {
				// TODO: 调用通知器发送告警
				fmt.Printf("[CPU Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *CPUPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *CPUPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *CPUPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 获取 CPU 使用率
	cpuPercents, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu percent: %w", err)
	}

	// 获取 Load Average
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get load average: %w", err)
	}

	cpuPercent := cpuPercents[0]

	// 检查是否超过阈值
	if cpuPercent > p.threshold {
		// 生成事件 ID
		eventID := fmt.Sprintf("cpu_high_%d", time.Now().Unix())

		return &base.Event{
			ID:        eventID,
			Type:      "cpu_high",
			Severity:  getSeverity(cpuPercent),
			Timestamp: time.Now(),
			Target:    getHostname(),
			Metric:    "cpu_usage",
			Value:     cpuPercent,
			Threshold: p.threshold,
			Message:   fmt.Sprintf("CPU 使用率过高：%.1f%% (阈值：%.1f%%)", cpuPercent, p.threshold),
			Labels: map[string]string{
				"plugin":    "cpu",
				"hostname":  getHostname(),
				"severity":  getSeverity(cpuPercent),
			},
			RawData: map[string]interface{}{
				"cpu_percent":    cpuPercent,
				"load_1m":        loadAvg.Load1,
				"load_5m":        loadAvg.Load5,
				"load_15m":       loadAvg.Load15,
			},
		}, nil
	}

	return nil, nil
}

// getSeverity 根据 CPU 使用率获取严重程度
func getSeverity(cpuPercent float64) string {
	if cpuPercent >= 95 {
		return "critical"
	} else if cpuPercent >= 85 {
		return "warning"
	}
	return "info"
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
