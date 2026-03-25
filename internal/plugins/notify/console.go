/*
Console 通知器

将告警事件输出到终端，支持彩色显示
*/
package notify

import (
	"fmt"
	"os"
	"siriusec-mcp/internal/plugins/base"
	"time"
)

// Color 颜色代码
type Color string

const (
	ColorReset  Color = "\033[0m"
	ColorRed    Color = "\033[31m"
	ColorGreen  Color = "\033[32m"
	ColorYellow Color = "\033[33m"
	ColorBlue   Color = "\033[34m"
	ColorPurple Color = "\033[35m"
	ColorCyan   Color = "\033[36m"
)

// ConsoleNotifier Console 通知器
type ConsoleNotifier struct {
	enableColor bool // 是否启用颜色
	location    *time.Location // 时区
}

// Name 通知器名称
func (n *ConsoleNotifier) Name() string {
	return "console"
}

// Init 初始化
func (n *ConsoleNotifier) Init(enableColor bool) {
	n.enableColor = enableColor
	
	// 检测是否支持颜色
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		n.enableColor = false
	}
	
	n.location = time.Local
}

// Send 发送通知
func (n *ConsoleNotifier) Send(event *base.Event) error {
	// 获取颜色
	color := n.getColorBySeverity(event.Severity)
	
	// 格式化时间
	timeStr := event.Timestamp.In(n.location).Format("2006-01-02 15:04:05")
	
	// 构建消息
	if n.enableColor {
		fmt.Printf("%s[%s] [%s] %s: %s%s\n",
			color,
			event.Severity,
			timeStr,
			event.Type,
			event.Message,
			ColorReset)
	} else {
		fmt.Printf("[%s] [%s] %s: %s\n",
			event.Severity,
			timeStr,
			event.Type,
			event.Message)
	}
	
	// 打印详细信息
	if len(event.Labels) > 0 {
		fmt.Printf("  Labels: %v\n", event.Labels)
	}
	
	if event.RawData != nil && len(event.RawData) > 0 {
		fmt.Printf("  Data: %v\n", event.RawData)
	}
	
	fmt.Println()
	
	return nil
}

// Close 关闭连接
func (n *ConsoleNotifier) Close() error {
	return nil
}

// getColorBySeverity 根据严重程度获取颜色
func (n *ConsoleNotifier) getColorBySeverity(severity string) Color {
	switch severity {
	case "critical":
		return ColorRed
	case "warning":
		return ColorYellow
	case "info":
		return ColorGreen
	default:
		return ColorReset
	}
}
package notify
