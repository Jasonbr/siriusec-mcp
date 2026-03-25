package disk
/*
磁盘监控插件

监控磁盘空间、Inode 使用率、IO 等待等
*/
package disk

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"siriusec-mcp/internal/plugins/base"
)

// DiskPlugin 磁盘监控插件
type DiskPlugin struct {
	config        base.PluginConfig
	status        base.PluginStatus
	usageThreshold float64 // 磁盘使用率阈值（百分比）
	inodeThreshold float64 // Inode 使用率阈值（百分比）
	mountPoint    string  // 挂载点，默认 /
}

// Name 插件名称
func (p *DiskPlugin) Name() string {
	return "disk"
}

// Description 插件描述
func (p *DiskPlugin) Description() string {
	return "磁盘空间、Inode 监控"
}

// Init 初始化插件
func (p *DiskPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取参数
	if threshold, ok := config.Params["usage_threshold"].(float64); ok {
		p.usageThreshold = threshold
	} else {
		p.usageThreshold = 85.0
	}

	if threshold, ok := config.Params["inode_threshold"].(float64); ok {
		p.inodeThreshold = threshold
	} else {
		p.inodeThreshold = 90.0
	}

	if mp, ok := config.Params["mount_point"].(string); ok {
		p.mountPoint = mp
	} else {
		p.mountPoint = "/"
	}

	return nil
}

// Start 启动监控循环
func (p *DiskPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("disk check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[DISK Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *DiskPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *DiskPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *DiskPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 获取磁盘使用情况
	usage, err := disk.UsageWithContext(ctx, p.mountPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	// 检查磁盘使用率
	if usage.UsedPercent > p.usageThreshold {
		return p.createDiskFullEvent(usage), nil
	}

	// 检查 Inode 使用率
	if usage.InodesUsedPercent > p.inodeThreshold {
		return p.createInodeFullEvent(usage), nil
	}

	return nil, nil
}

// createDiskFullEvent 创建磁盘空间不足事件
func (p *DiskPlugin) createDiskFullEvent(usage *disk.UsageStat) *base.Event {
	severity := "warning"
	if usage.UsedPercent >= 95 {
		severity = "critical"
	}

	return &base.Event{
		ID:        fmt.Sprintf("disk_full_%d", time.Now().Unix()),
		Type:      "disk_full",
		Severity:  severity,
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "disk_usage",
		Value:     usage.UsedPercent,
		Threshold: p.usageThreshold,
		Message:   fmt.Sprintf("磁盘空间不足：%s 已用 %.1f%% (阈值：%.1f%%)", p.mountPoint, usage.UsedPercent, p.usageThreshold),
		Labels: map[string]string{
			"plugin":    "disk",
			"hostname":  getHostname(),
			"severity":  severity,
			"mount":     p.mountPoint,
		},
		RawData: map[string]interface{}{
			"total":       usage.Total,
			"used":        usage.Used,
			"free":        usage.Free,
			"used_percent": usage.UsedPercent,
			"inodes_total": usage.InodesTotal,
			"inodes_used":  usage.InodesUsed,
			"inodes_free":  usage.InodesFree,
			"inodes_used_percent": usage.InodesUsedPercent,
			"fstype":      usage.Fstype,
		},
	}
}

// createInodeFullEvent 创建 Inode 不足事件
func (p *DiskPlugin) createInodeFullEvent(usage *disk.UsageStat) *base.Event {
	severity := "warning"
	if usage.InodesUsedPercent >= 95 {
		severity = "critical"
	}

	return &base.Event{
		ID:        fmt.Sprintf("inode_full_%d", time.Now().Unix()),
		Type:      "inode_full",
		Severity:  severity,
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "inode_usage",
		Value:     usage.InodesUsedPercent,
		Threshold: p.inodeThreshold,
		Message:   fmt.Sprintf("Inode 不足：%s 已用 %.1f%% (阈值：%.1f%%)", p.mountPoint, usage.InodesUsedPercent, p.inodeThreshold),
		Labels: map[string]string{
			"plugin":    "disk",
			"hostname":  getHostname(),
			"severity":  severity,
			"mount":     p.mountPoint,
		},
		RawData: map[string]interface{}{
			"inodes_total": usage.InodesTotal,
			"inodes_used":  usage.InodesUsed,
			"inodes_free":  usage.InodesFree,
			"inodes_used_percent": usage.InodesUsedPercent,
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
