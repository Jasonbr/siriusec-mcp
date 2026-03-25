package process
/*
进程监控插件

监控进程数量、僵尸进程、Top N 资源占用进程等
*/
package process

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"siriusec-mcp/internal/plugins/base"
)

// ProcessPlugin 进程监控插件
type ProcessPlugin struct {
	config         base.PluginConfig
	status         base.PluginStatus
	zombieThreshold int     // 僵尸进程阈值
	topCPUPercent  float64 // Top CPU 占用阈值
	topMemPercent  float64 // Top 内存占用阈值
}

// Name 插件名称
func (p *ProcessPlugin) Name() string {
	return "process"
}

// Description 插件描述
func (p *ProcessPlugin) Description() string {
	return "进程数量、僵尸进程、资源占用监控"
}

// Init 初始化插件
func (p *ProcessPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取阈值
	if threshold, ok := config.Params["zombie_threshold"].(float64); ok {
		p.zombieThreshold = int(threshold)
	} else {
		p.zombieThreshold = 10
	}

	if percent, ok := config.Params["top_cpu_percent"].(float64); ok {
		p.topCPUPercent = percent
	} else {
		p.topCPUPercent = 80.0
	}

	if percent, ok := config.Params["top_mem_percent"].(float64); ok {
		p.topMemPercent = percent
	} else {
		p.topMemPercent = 80.0
	}

	return nil
}

// Start 启动监控循环
func (p *ProcessPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("process check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[PROCESS Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *ProcessPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *ProcessPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *ProcessPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 获取所有进程
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	// 检查僵尸进程
	zombieEvent := p.checkZombieProcesses(processes)
	if zombieEvent != nil {
		return zombieEvent, nil
	}

	// 检查高 CPU 占用进程
	cpuEvent := p.checkHighCPUProcesses(processes)
	if cpuEvent != nil {
		return cpuEvent, nil
	}

	// 检查高内存占用进程
	memEvent := p.checkHighMemProcesses(processes)
	if memEvent != nil {
		return memEvent, nil
	}

	return nil, nil
}

// checkZombieProcesses 检查僵尸进程
func (p *ProcessPlugin) checkZombieProcesses(processes []*process.Process) *base.Event {
	zombieCount := 0
	var zombiePids []int32

	for _, p := range processes {
		status, err := p.Status()
		if err != nil {
			continue
		}

		for _, s := range status {
			if s == "zombie" {
				zombieCount++
				zombiePids = append(zombiePids, p.Pid)
			}
		}
	}

	if zombieCount > p.zombieThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("zombie_high_%d", time.Now().Unix()),
			Type:      "zombie_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    getHostname(),
			Metric:    "zombie_processes",
			Value:     zombieCount,
			Threshold: float64(p.zombieThreshold),
			Message:   fmt.Sprintf("僵尸进程过多：%d 个 (阈值：%d)", zombieCount, p.zombieThreshold),
			Labels: map[string]string{
				"plugin":   "process",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"zombie_count":  zombieCount,
				"zombie_pids":   zombiePids,
				"total_process": len(processes),
			},
		}
	}

	return nil
}

// checkHighCPUProcesses 检查高 CPU 占用进程
func (p *ProcessPlugin) checkHighCPUProcesses(processes []*process.Process) *base.Event {
	var highCPUProcs []map[string]interface{}

	for _, p := range processes {
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		if cpuPercent > p.topCPUPercent {
			name, _ := p.Name()
			highCPUProcs = append(highCPUProcs, map[string]interface{}{
				"name":      name,
				"pid":       p.Pid,
				"cpu":       cpuPercent,
				"threads":   getThreadCount(p),
			})
		}
	}

	if len(highCPUProcs) > 0 {
		return &base.Event{
			ID:        fmt.Sprintf("high_cpu_%d", time.Now().Unix()),
			Type:      "high_cpu_process",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    getHostname(),
			Metric:    "high_cpu_processes",
			Value:     len(highCPUProcs),
			Threshold: 1.0,
			Message:   fmt.Sprintf("发现 %d 个高 CPU 占用进程", len(highCPUProcs)),
			Labels: map[string]string{
				"plugin":   "process",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"high_cpu_processes": highCPUProcs,
				"threshold":          p.topCPUPercent,
			},
		}
	}

	return nil
}

// checkHighMemProcesses 检查高内存占用进程
func (p *ProcessPlugin) checkHighMemProcesses(processes []*process.Process) *base.Event {
	var highMemProcs []map[string]interface{}

	for _, p := range processes {
		memPercent, err := p.MemoryPercent()
		if err != nil {
			continue
		}

		if memPercent > p.topMemPercent {
			name, _ := p.Name()
			highMemProcs = append(highMemProcs, map[string]interface{}{
				"name":      name,
				"pid":       p.Pid,
				"memory":    memPercent,
				"memory_info": getMemoryInfo(p),
			})
		}
	}

	if len(highMemProcs) > 0 {
		return &base.Event{
			ID:        fmt.Sprintf("high_mem_%d", time.Now().Unix()),
			Type:      "high_mem_process",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    getHostname(),
			Metric:    "high_memory_processes",
			Value:     len(highMemProcs),
			Threshold: 1.0,
			Message:   fmt.Sprintf("发现 %d 个高内存占用进程", len(highMemProcs)),
			Labels: map[string]string{
				"plugin":   "process",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"high_mem_processes": highMemProcs,
				"threshold":          p.topMemPercent,
			},
		}
	}

	return nil
}

// getThreadCount 获取线程数
func getThreadCount(p *process.Process) int32 {
	threads, _ := p.NumThreads()
	return threads
}

// getMemoryInfo 获取内存详情
func getMemoryInfo(p *process.Process) map[string]interface{} {
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"rss":  memInfo.RSS,
		"vms":  memInfo.VMS,
		"swap": memInfo.Swap,
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
