package mysql
/*
MySQL 监控插件

监控 MySQL 连接数、慢查询、主从延迟、锁等待等
*/
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"siriusec-mcp/internal/plugins/base"
)

// MySQLPlugin MySQL 监控插件
type MySQLPlugin struct {
	config          base.PluginConfig
	status          base.PluginStatus
	dsn             string  // Data Source Name
	connThreshold   int     // 连接数阈值
	slowQueryThreshold int   // 慢查询阈值
	replDelayThreshold int  // 主从延迟阈值（秒）
}

// Name 插件名称
func (p *MySQLPlugin) Name() string {
	return "mysql"
}

// Description 插件描述
func (p *MySQLPlugin) Description() string {
	return "MySQL 连接数、慢查询、主从延迟、锁等待监控"
}

// Init 初始化插件
func (p *MySQLPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取参数
	if dsn, ok := config.Params["dsn"].(string); ok {
		p.dsn = dsn
	} else {
		p.dsn = "root@tcp(localhost:3306)/?timeout=5s"
	}

	if threshold, ok := config.Params["conn_threshold"].(float64); ok {
		p.connThreshold = int(threshold)
	} else {
		p.connThreshold = 500
	}

	if threshold, ok := config.Params["slow_query_threshold"].(float64); ok {
		p.slowQueryThreshold = int(threshold)
	} else {
		p.slowQueryThreshold = 100
	}

	if threshold, ok := config.Params["repl_delay_threshold"].(float64); ok {
		p.replDelayThreshold = int(threshold)
	} else {
		p.replDelayThreshold = 60
	}

	return nil
}

// Start 启动监控循环
func (p *MySQLPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("mysql check failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[MYSQL Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *MySQLPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *MySQLPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *MySQLPlugin) Check(ctx context.Context) (*base.Event, error) {
	// 创建数据库连接
	db, err := sql.Open("mysql", p.dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// 设置连接超时
	db.SetConnMaxLifetime(time.Second * 10)

	// 检查连接性
	if err := db.PingContext(ctx); err != nil {
		return &base.Event{
			ID:        fmt.Sprintf("mysql_down_%d", time.Now().Unix()),
			Type:      "mysql_unreachable",
			Severity:  "critical",
			Timestamp: time.Now(),
			Target:    p.dsn,
			Metric:    "connection_status",
			Value:     "down",
			Threshold: "up",
			Message:   fmt.Sprintf("MySQL 无法连接：%v", err),
			Labels: map[string]string{
				"plugin":   "mysql",
				"hostname": getHostname(),
				"severity": "critical",
			},
			RawData: map[string]interface{}{
				"error": err.Error(),
			},
		}, nil
	}

	// 检查连接数
	if connEvent := p.checkConnections(db); connEvent != nil {
		return connEvent, nil
	}

	// 检查慢查询
	if slowEvent := p.checkSlowQueries(db); slowEvent != nil {
		return slowEvent, nil
	}

	// 检查主从延迟
	if replEvent := p.checkReplicationLag(db); replEvent != nil {
		return replEvent, nil
	}

	// 检查锁等待
	if lockEvent := p.checkLockWait(db); lockEvent != nil {
		return lockEvent, nil
	}

	return nil, nil
}

// checkConnections 检查连接数
func (p *MySQLPlugin) checkConnections(db *sql.DB) *base.Event {
	var status struct {
		Name  string
		Value string
	}

	err := db.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(&status.Name, &status.Value)
	if err != nil {
		return nil
	}

	connected, _ := strconv.Atoi(status.Value)

	if connected > p.connThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("mysql_conn_high_%d", time.Now().Unix()),
			Type:      "mysql_connections_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.dsn,
			Metric:    "threads_connected",
			Value:     connected,
			Threshold: float64(p.connThreshold),
			Message:   fmt.Sprintf("MySQL 连接数过多：%d", connected),
			Labels: map[string]string{
				"plugin":   "mysql",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"threads_connected": connected,
				"max_connections":   p.getMaxConnections(db),
			},
		}
	}

	return nil
}

// checkSlowQueries 检查慢查询
func (p *MySQLPlugin) checkSlowQueries(db *sql.DB) *base.Event {
	var status struct {
		Name  string
		Value string
	}

	err := db.QueryRow("SHOW GLOBAL STATUS LIKE 'Slow_queries'").Scan(&status.Name, &status.Value)
	if err != nil {
		return nil
	}

	slowQueries, _ := strconv.Atoi(status.Value)

	if slowQueries > p.slowQueryThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("mysql_slow_high_%d", time.Now().Unix()),
			Type:      "mysql_slow_queries_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.dsn,
			Metric:    "slow_queries",
			Value:     slowQueries,
			Threshold: float64(p.slowQueryThreshold),
			Message:   fmt.Sprintf("MySQL 慢查询过多：%d", slowQueries),
			Labels: map[string]string{
				"plugin":   "mysql",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"slow_queries":       slowQueries,
				"long_query_time":    p.getLongQueryTime(db),
				"questions_per_sec":  p.getQuestionsPerSec(db),
			},
		}
	}

	return nil
}

// checkReplicationLag 检查主从延迟
func (p *MySQLPlugin) checkReplicationLag(db *sql.DB) *base.Event {
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return nil
	}
	defer rows.Close()

	// 获取列索引
	columns, _ := rows.Columns()
	slaveIoRunningIdx := -1
	slaveSqlRunningIdx := -1
	secondsBehindMasterIdx := -1

	for i, col := range columns {
		switch strings.ToLower(col) {
		case "slave_io_running":
			slaveIoRunningIdx = i
		case "slave_sql_running":
			slaveSqlRunningIdx = i
		case "seconds_behind_master":
			secondsBehindMasterIdx = i
		}
	}

	if secondsBehindMasterIdx == -1 {
		return nil
	}

	// 解析结果
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// 检查复制状态
		if slaveIoRunningIdx >= 0 && values[slaveIoRunningIdx].String == "No" {
			return &base.Event{
				ID:        fmt.Sprintf("mysql_slave_stopped_%d", time.Now().Unix()),
				Type:      "mysql_replication_stopped",
				Severity:  "critical",
				Timestamp: time.Now(),
				Target:    p.dsn,
				Metric:    "slave_io_running",
				Value:     "No",
				Threshold: "Yes",
				Message:   "MySQL 从库复制已停止",
				Labels: map[string]string{
					"plugin":   "mysql",
					"hostname": getHostname(),
					"severity": "critical",
				},
				RawData: map[string]interface{}{
					"slave_io_running":  values[slaveIoRunningIdx].String,
					"slave_sql_running": values[slaveSqlRunningIdx].String,
				},
			}
		}

		// 检查延迟时间
		if secondsBehindMasterIdx >= 0 && values[secondsBehindMasterIdx].Valid {
			seconds, _ := strconv.Atoi(values[secondsBehindMasterIdx].String)
			if seconds > p.replDelayThreshold {
				return &base.Event{
					ID:        fmt.Sprintf("mysql_repl_lag_%d", time.Now().Unix()),
					Type:      "mysql_replication_lag",
					Severity:  "warning",
					Timestamp: time.Now(),
					Target:    p.dsn,
					Metric:    "seconds_behind_master",
					Value:     seconds,
					Threshold: float64(p.replDelayThreshold),
					Message:   fmt.Sprintf("MySQL 主从延迟过高：%d秒", seconds),
					Labels: map[string]string{
						"plugin":   "mysql",
						"hostname": getHostname(),
						"severity": "warning",
					},
					RawData: map[string]interface{}{
						"seconds_behind_master": seconds,
						"slave_io_running":      values[slaveIoRunningIdx].String,
						"slave_sql_running":     values[slaveSqlRunningIdx].String,
					},
				}
			}
		}
	}

	return nil
}

// checkLockWait 检查锁等待
func (p *MySQLPlugin) checkLockWait(db *sql.DB) *base.Event {
	var status struct {
		Name  string
		Value string
	}

	err := db.QueryRow("SHOW GLOBAL STATUS LIKE 'Innodb_row_lock_current_waits'").Scan(&status.Name, &status.Value)
	if err != nil {
		return nil
	}

	currentWaits, _ := strconv.Atoi(status.Value)

	if currentWaits > 10 {
		return &base.Event{
			ID:        fmt.Sprintf("mysql_lock_wait_%d", time.Now().Unix()),
			Type:      "mysql_lock_wait_high",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.dsn,
			Metric:    "innodb_row_lock_waits",
			Value:     currentWaits,
			Threshold: 10.0,
			Message:   fmt.Sprintf("MySQL 锁等待过多：%d", currentWaits),
			Labels: map[string]string{
				"plugin":   "mysql",
				"hostname": getHostname(),
				"severity": "warning",
			},
			RawData: map[string]interface{}{
				"innodb_row_lock_current_waits": currentWaits,
			},
		}
	}

	return nil
}

// getMaxConnections 获取最大连接数
func (p *MySQLPlugin) getMaxConnections(db *sql.DB) int {
	var value string
	err := db.QueryRow("SHOW VARIABLES LIKE 'max_connections'").Scan(new(string), &value)
	if err != nil {
		return 0
	}
	maxConn, _ := strconv.Atoi(value)
	return maxConn
}

// getLongQueryTime 获取慢查询阈值
func (p *MySQLPlugin) getLongQueryTime(db *sql.DB) float64 {
	var value string
	err := db.QueryRow("SHOW VARIABLES LIKE 'long_query_time'").Scan(new(string), &value)
	if err != nil {
		return 0
	}
	lqt, _ := strconv.ParseFloat(value, 64)
	return lqt
}

// getQuestionsPerSec 获取每秒查询数
func (p *MySQLPlugin) getQuestionsPerSec(db *sql.DB) float64 {
	var status struct {
		Name  string
		Value string
	}

	err := db.QueryRow("SHOW GLOBAL STATUS LIKE 'Questions'").Scan(&status.Name, &status.Value)
	if err != nil {
		return 0
	}

	questions, _ := strconv.Atoi(status.Value)
	// 这里简化处理，假设运行了 1 秒
	return float64(questions)
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
