package redis
/*
Redis 监控插件

监控 Redis 内存、连接数、命中率、主从状态等
*/
package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"siriusec-mcp/internal/plugins/base"
)

// RedisPlugin Redis 监控插件
type RedisPlugin struct {
	config        base.PluginConfig
	status        base.PluginStatus
	address       string  // Redis 地址
	password      string  // Redis 密码
	memoryThreshold float64 // 内存使用率阈值
	connThreshold   int     // 连接数阈值
	hitRateThreshold float64 // 命中率阈值（百分比）
}

// Name 插件名称
func (p *RedisPlugin) Name() string {
	return "redis"
}

// Description 插件描述
func (p *RedisPlugin) Description() string {
	return "Redis 内存、连接数、命中率、主从状态监控"
}

// Init 初始化插件
func (p *RedisPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取参数
	if address, ok := config.Params["address"].(string); ok {
		p.address = address
	} else {
		p.address = "localhost:6379"
	}

	if password, ok := config.Params["password"].(string); ok {
		p.password = password
	}

	if threshold, ok := config.Params["memory_threshold"].(float64); ok {
		p.memoryThreshold = threshold
	} else {
		p.memoryThreshold = 85.0
	}

	if threshold, ok := config.Params["conn_threshold"].(float64); ok {
		p.connThreshold = int(threshold)
	} else {
		p.connThreshold = 10000
	}

	if threshold, ok := config.Params["hitrate_threshold"].(float64); ok {
		p.hitRateThreshold = threshold
	} else {
		p.hitRateThreshold = 80.0
	}

	return nil
}

// Start 启动监控循环
func (p *RedisPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("redis check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[REDIS Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *RedisPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *RedisPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *RedisPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     p.address,
		Password: p.password,
		DB:       0,
	})

	// 获取 Redis 信息
	info, err := client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", err)
	}

	stats := p.parseRedisInfo(info)

	// 检查内存使用率
	if memEvent := p.checkMemory(stats); memEvent != nil {
		return memEvent, nil
	}

	// 检查连接数
	if connEvent := p.checkConnections(stats); connEvent != nil {
		return connEvent, nil
	}

	// 检查命中率
	if hitEvent := p.checkHitRate(stats); hitEvent != nil {
		return hitEvent, nil
	}

	// 检查主从同步
	if slaveEvent := p.checkReplication(stats); slaveEvent != nil {
		return slaveEvent, nil
	}

	return nil, nil
}

// checkMemory 检查内存使用
func (p *RedisPlugin) checkMemory(stats map[string]string) *base.Event {
	usedMemory, _ := strconv.ParseFloat(stats["used_memory"], 64)
	maxMemory, _ := strconv.ParseFloat(stats["maxmemory"], 64)

	var usedPercent float64
	if maxMemory > 0 {
		usedPercent = usedMemory / maxMemory * 100
	} else {
		// 如果没有设置 maxmemory，使用系统内存作为参考
		// 这里简化处理，假设最大 8GB
		usedPercent = usedMemory / (8 * 1024 * 1024 * 1024) * 100
	}

	if usedPercent > p.memoryThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("redis_mem_high_%d", time.Now().Unix()),
			Type:      "redis_memory_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.address,
			Metric:    "memory_usage_percent",
			Value:     usedPercent,
			Threshold: p.memoryThreshold,
			Message:   fmt.Sprintf("Redis 内存使用率过高：%.1f%%", usedPercent),
			Labels: map[string]string{
				"plugin":   "redis",
				"hostname": getHostname(),
				"severity": "warning",
				"address":  p.address,
			},
			RawData: map[string]interface{}{
				"used_memory":      formatBytes(usedMemory),
				"max_memory":       formatBytes(maxMemory),
				"used_percent":     usedPercent,
				"mem_fragmentation_ratio": stats["mem_fragmentation_ratio"],
			},
		}
	}

	return nil
}

// checkConnections 检查连接数
func (p *RedisPlugin) checkConnections(stats map[string]string) *base.Event {
	connectedClients, _ := strconv.Atoi(stats["connected_clients"])

	if connectedClients > p.connThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("redis_conn_high_%d", time.Now().Unix()),
			Type:      "redis_connections_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.address,
			Metric:    "connected_clients",
			Value:     connectedClients,
			Threshold: float64(p.connThreshold),
			Message:   fmt.Sprintf("Redis 连接数过多：%d", connectedClients),
			Labels: map[string]string{
				"plugin":   "redis",
				"hostname": getHostname(),
				"severity": "warning",
				"address":  p.address,
			},
			RawData: map[string]interface{}{
				"connected_clients": connectedClients,
				"blocked_clients":   stats["blocked_clients"],
				"rejected_connections": stats["rejected_connections"],
			},
		}
	}

	return nil
}

// checkHitRate 检查命中率
func (p *RedisPlugin) checkHitRate(stats map[string]string) *base.Event {
	keyspaceHits, _ := strconv.ParseFloat(stats["keyspace_hits"], 64)
	keyspaceMisses, _ := strconv.ParseFloat(stats["keyspace_misses"], 64)

	total := keyspaceHits + keyspaceMisses
	var hitRate float64
	if total > 0 {
		hitRate = keyspaceHits / total * 100
	}

	if hitRate < p.hitRateThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("redis_hitrate_low_%d", time.Now().Unix()),
			Type:      "redis_hit_rate_low",
			Severity:  "info",
			Timestamp: time.Now(),
			Target:    p.address,
			Metric:    "hit_rate_percent",
			Value:     hitRate,
			Threshold: p.hitRateThreshold,
			Message:   fmt.Sprintf("Redis 命中率过低：%.1f%%", hitRate),
			Labels: map[string]string{
				"plugin":   "redis",
				"hostname": getHostname(),
				"severity": "info",
				"address":  p.address,
			},
			RawData: map[string]interface{}{
				"keyspace_hits":   keyspaceHits,
				"keyspace_misses": keyspaceMisses,
				"hit_rate":        hitRate,
			},
		}
	}

	return nil
}

// checkReplication 检查主从同步
func (p *RedisPlugin) checkReplication(stats map[string]string) *base.Event {
	role := stats["role"]
	
	if role == "slave" {
		masterLinkStatus := stats["master_link_status"]
		if masterLinkStatus == "down" {
			return &base.Event{
				ID:        fmt.Sprintf("redis_slave_down_%d", time.Now().Unix()),
				Type:      "redis_slave_disconnected",
				Severity:  "critical",
				Timestamp: time.Now(),
				Target:    p.address,
				Metric:    "master_link_status",
				Value:     "down",
				Threshold: "up",
				Message:   "Redis 从节点与主节点断开连接",
				Labels: map[string]string{
					"plugin":   "redis",
					"hostname": getHostname(),
					"severity": "critical",
					"address":  p.address,
					"role":     "slave",
				},
				RawData: map[string]interface{}{
					"role":                 role,
					"master_link_status":   masterLinkStatus,
					"master_last_io_seconds_ago": stats["master_last_io_seconds_ago"],
				},
			}
		}

		// 检查同步延迟
		masterLastIoSeconds, _ := strconv.Atoi(stats["master_last_io_seconds_ago"])
		if masterLastIoSeconds > 60 {
			return &base.Event{
				ID:        fmt.Sprintf("redis_sync_lag_%d", time.Now().Unix()),
				Type:      "redis_sync_lag",
				Severity:  "warning",
				Timestamp: time.Now(),
				Target:    p.address,
				Metric:    "master_last_io_seconds",
				Value:     masterLastIoSeconds,
				Threshold: 60,
				Message:   fmt.Sprintf("Redis 主从同步延迟：%d秒", masterLastIoSeconds),
				Labels: map[string]string{
					"plugin":   "redis",
					"hostname": getHostname(),
					"severity": "warning",
					"address":  p.address,
					"role":     "slave",
				},
				RawData: map[string]interface{}{
					"master_last_io_seconds": masterLastIoSeconds,
					"sync_full":              stats["sync_full"],
					"sync_partial_ok":        stats["sync_partial_ok"],
				},
			}
		}
	}

	return nil
}

// parseRedisInfo 解析 Redis INFO 命令输出
func (p *RedisPlugin) parseRedisInfo(info string) map[string]string {
	stats := make(map[string]string)
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			stats[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return stats
}

// formatBytes 格式化字节数
func formatBytes(bytes float64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", bytes/GB)
	} else if bytes >= MB {
		return fmt.Sprintf("%.2f MB", bytes/MB)
	} else if bytes >= KB {
		return fmt.Sprintf("%.2f KB", bytes/KB)
	}
	return fmt.Sprintf("%.0f B", bytes)
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
