package net
/*
网络监控插件

监控网络接口状态、流量、丢包率、TCP 连接等
*/
package net

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/net"
	"siriusec-mcp/internal/plugins/base"
)

// NetPlugin 网络监控插件
type NetPlugin struct {
	config        base.PluginConfig
	status        base.PluginStatus
	dropThreshold float64 // 丢包率阈值（百分比）
	errThreshold  float64 // 错误包阈值（百分比）
}

// Name 插件名称
func (p *NetPlugin) Name() string {
	return "net"
}

// Description 插件描述
func (p *NetPlugin) Description() string {
	return "网络接口、流量、丢包监控"
}

// Init 初始化插件
func (p *NetPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取阈值，默认丢包率 1%，错误包 0.1%
	if threshold, ok := config.Params["drop_threshold"].(float64); ok {
		p.dropThreshold = threshold
	} else {
		p.dropThreshold = 1.0
	}

	if threshold, ok := config.Params["err_threshold"].(float64); ok {
		p.errThreshold = threshold
	} else {
		p.errThreshold = 0.1
	}

	return nil
}

// Start 启动监控循环
func (p *NetPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("net check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[NET Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *NetPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *NetPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *NetPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 获取网络接口统计
	ioCounters, err := net.IOCountersWithContext(ctx, true) // true = per interface
	if err != nil {
		return nil, fmt.Errorf("failed to get net IO counters: %w", err)
	}

	// 检查每个网络接口
	for _, io := range ioCounters {
		event := p.checkInterface(io)
		if event != nil {
			return event, nil
		}
	}

	// 检查 TCP 连接数
	tcpEvent, err := p.checkTCPConnections(ctx)
	if err != nil {
		return nil, err
	}
	if tcpEvent != nil {
		return tcpEvent, nil
	}

	return nil, nil
}

// checkInterface 检查单个网络接口
func (p *NetPlugin) checkInterface(io net.IOCountersStat) *base.Event {
	// 计算丢包率
	totalPackets := io.PacketsSent + io.PacketsRecv
	totalDrops := io.Dropin + io.Dropout
	
	var dropPercent float64
	if totalPackets > 0 {
		dropPercent = float64(totalDrops) / float64(totalPackets) * 100
	}

	// 计算错误包率
	totalErrors := io.Errin + io.Errout
	var errPercent float64
	if totalPackets > 0 {
		errPercent = float64(totalErrors) / float64(totalPackets) * 100
	}

	// 检查丢包率
	if dropPercent > p.dropThreshold {
		return p.createDropHighEvent(io, dropPercent)
	}

	// 检查错误包率
	if errPercent > p.errThreshold {
		return p.createErrHighEvent(io, errPercent)
	}

	return nil
}

// createDropHighEvent 创建丢包过高事件
func (p *NetPlugin) createDropHighEvent(io net.IOCountersStat, dropPercent float64) *base.Event {
	severity := "warning"
	if dropPercent >= 5 {
		severity = "critical"
	}

	return &base.Event{
		ID:        fmt.Sprintf("net_drop_%s_%d", io.Name, time.Now().Unix()),
		Type:      "net_drop_high",
		Severity:  severity,
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "packet_drop_rate",
		Value:     dropPercent,
		Threshold: p.dropThreshold,
		Message:   fmt.Sprintf("网络接口 %s 丢包率过高：%.2f%% (阈值：%.1f%%)", io.Name, dropPercent, p.dropThreshold),
		Labels: map[string]string{
			"plugin":    "net",
			"hostname":  getHostname(),
			"severity":  severity,
			"interface": io.Name,
		},
		RawData: map[string]interface{}{
			"name":            io.Name,
			"bytes_sent":      io.BytesSent,
			"bytes_recv":      io.BytesRecv,
			"packets_sent":    io.PacketsSent,
			"packets_recv":    io.PacketsRecv,
			"drop_in":         io.Dropin,
			"drop_out":        io.Dropout,
			"drop_percent":    dropPercent,
		},
	}
}

// createErrHighEvent 创建错误包过高事件
func (p *NetPlugin) createErrHighEvent(io net.IOCountersStat, errPercent float64) *base.Event {
	severity := "warning"
	if errPercent >= 1 {
		severity = "critical"
	}

	return &base.Event{
		ID:        fmt.Sprintf("net_err_%s_%d", io.Name, time.Now().Unix()),
		Type:      "net_err_high",
		Severity:  severity,
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "packet_error_rate",
		Value:     errPercent,
		Threshold: p.errThreshold,
		Message:   fmt.Sprintf("网络接口 %s 错误包过高：%.2f%% (阈值：%.1f%%)", io.Name, errPercent, p.errThreshold),
		Labels: map[string]string{
			"plugin":    "net",
			"hostname":  getHostname(),
			"severity":  severity,
			"interface": io.Name,
		},
		RawData: map[string]interface{}{
			"name":            io.Name,
			"bytes_sent":      io.BytesSent,
			"bytes_recv":      io.BytesRecv,
			"packets_sent":    io.PacketsSent,
			"packets_recv":    io.PacketsRecv,
			"err_in":          io.Errin,
			"err_out":         io.Errout,
			"error_percent":   errPercent,
		},
	}
}

// checkTCPConnections 检查 TCP 连接
func (p *NetPlugin) checkTCPConnections(ctx context.Context) (*base.Event, error) {
	// 获取所有 TCP 连接
	connections, err := net.ConnectionsWithContext(ctx, "tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get TCP connections: %w", err)
	}

	// 统计各状态连接数
	stateCount := make(map[string]int)
	for _, conn := range connections {
		stateCount[conn.Status]++
	}

	// 检查 TIME_WAIT 数量
	timeWaitCount := stateCount["TIME_WAIT"]
	if timeWaitCount > 1000 {
		return &base.Event{
			ID:        fmt.Sprintf("tcp_timewait_%d", time.Now().Unix()),
			Type:      "tcp_timewait_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    getHostname(),
			Metric:    "tcp_time_wait",
			Value:     timeWaitCount,
			Threshold: 1000,
			Message:   fmt.Sprintf("TCP TIME_WAIT 连接过多：%d 个", timeWaitCount),
			Labels: map[string]string{
				"plugin":   "net",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"time_wait_count": timeWaitCount,
				"established":     stateCount["ESTABLISHED"],
				"listen":          stateCount["LISTEN"],
				"close_wait":      stateCount["CLOSE_WAIT"],
				"total":           len(connections),
			},
		}, nil
	}

	return nil, nil
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
