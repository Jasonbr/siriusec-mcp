/*
Docker 容器监控插件

监控容器状态、资源使用、频繁重启等
*/
package docker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"siriusec-mcp/internal/plugins/base"
)

// DockerPlugin Docker 容器监控插件
type DockerPlugin struct {
	config        base.PluginConfig
	status        base.PluginStatus
	endpoint      string  // Docker socket 路径
	restartThreshold int   // 重启次数阈值
}

// Name 插件名称
func (p *DockerPlugin) Name() string {
	return "docker"
}

// Description 插件描述
func (p *DockerPlugin) Description() string {
	return "Docker 容器状态和资源监控"
}

// Init 初始化插件
func (p *DockerPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取参数
	if endpoint, ok := config.Params["endpoint"].(string); ok {
		p.endpoint = endpoint
	} else {
		p.endpoint = "unix:///var/run/docker.sock"
	}

	if threshold, ok := config.Params["restart_threshold"].(float64); ok {
		p.restartThreshold = int(threshold)
	} else {
		p.restartThreshold = 3
	}

	return nil
}

// Start 启动监控循环
func (p *DockerPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("docker check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[DOCKER Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *DockerPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *DockerPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *DockerPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(
		client.WithHost(p.endpoint),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// 获取所有容器
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// 检查频繁重启的容器
	for _, container := range containers {
		// 检查重启次数
		if container.RestartCount > p.restartThreshold {
			return p.createFrequentRestartEvent(container), nil
		}

		// 检查容器健康状态
		if container.State == "running" && container.Status == "unhealthy" {
			return p.createUnhealthyEvent(container), nil
		}
	}

	// 检查已停止的重要容器
	stoppedContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: map[string][]string{
			"status": {"exited"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list stopped containers: %w", err)
	}

	// 如果有意外退出的容器，告警
	for _, container := range stoppedContainers {
		if isImportantContainer(container.Labels) {
			return p.createExitedEvent(container), nil
		}
	}

	return nil, nil
}

// createFrequentRestartEvent 创建频繁重启事件
func (p *DockerPlugin) createFrequentRestartEvent(container types.Container) *base.Event {
	return &base.Event{
		ID:        fmt.Sprintf("docker_restart_%s_%d", container.ID[:12], time.Now().Unix()),
		Type:      "docker_frequent_restart",
		Severity:  "warning",
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "restart_count",
		Value:     container.RestartCount,
		Threshold: float64(p.restartThreshold),
		Message:   fmt.Sprintf("容器 %s 频繁重启：%d 次", container.Names[0], container.RestartCount),
		Labels: map[string]string{
			"plugin":      "docker",
			"hostname":    getHostname(),
			"severity":    "warning",
			"container":   container.Names[0],
			"image":       container.Image,
			"status":      container.State,
		},
		RawData: map[string]interface{}{
			"container_id":   container.ID[:12],
			"container_name": container.Names[0],
			"image":          container.Image,
			"restart_count":  container.RestartCount,
			"status":         container.State,
			"created":        container.Created,
		},
	}
}

// createUnhealthyEvent 创建不健康事件
func (p *DockerPlugin) createUnhealthyEvent(container types.Container) *base.Event {
	return &base.Event{
		ID:        fmt.Sprintf("docker_unhealthy_%s_%d", container.ID[:12], time.Now().Unix()),
		Type:      "docker_unhealthy",
		Severity:  "warning",
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "health_status",
		Value:     "unhealthy",
		Threshold: "healthy",
		Message:   fmt.Sprintf("容器 %s 健康检查失败", container.Names[0]),
		Labels: map[string]string{
			"plugin":      "docker",
			"hostname":    getHostname(),
			"severity":    "warning",
			"container":   container.Names[0],
			"image":       container.Image,
		},
		RawData: map[string]interface{}{
			"container_id":   container.ID[:12],
			"container_name": container.Names[0],
			"image":          container.Image,
			"health":         container.Status,
		},
	}
}

// createExitedEvent 创建退出事件
func (p *DockerPlugin) createExitedEvent(container types.Container) *base.Event {
	return &base.Event{
		ID:        fmt.Sprintf("docker_exited_%s_%d", container.ID[:12], time.Now().Unix()),
		Type:      "docker_exited",
		Severity:  "critical",
		Timestamp: time.Now(),
		Target:    getHostname(),
		Metric:    "container_status",
		Value:     "exited",
		Threshold: "running",
		Message:   fmt.Sprintf("重要容器 %s 已退出", container.Names[0]),
		Labels: map[string]string{
			"plugin":      "docker",
			"hostname":    getHostname(),
			"severity":    "critical",
			"container":   container.Names[0],
			"image":       container.Image,
		},
		RawData: map[string]interface{}{
			"container_id":   container.ID[:12],
			"container_name": container.Names[0],
			"image":          container.Image,
			"state":          container.State,
		},
	}
}

// isImportantContainer 判断是否是重要容器
func isImportantContainer(labels map[string]string) bool {
	// 检查是否有 important=true 标签
	if important, ok := labels["important"]; ok && important == "true" {
		return true
	}

	// 检查是否有 restart_policy=always
	if policy, ok := labels["restart_policy"]; ok && policy == "always" {
		return true
	}

	return false
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
package docker
