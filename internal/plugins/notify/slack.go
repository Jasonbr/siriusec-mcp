/*
Slack 通知器

通过 Slack Webhook 发送告警消息
*/
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"siriusec-mcp/internal/plugins/base"
	"time"
)

// SlackNotifier Slack 通知器
type SlackNotifier struct {
	webhook   string            // Slack Incoming Webhook URL
	channel   string            // 频道（可选，覆盖默认）
	username  string            // 用户名（可选）
	iconEmoji string            // Emoji 图标（可选）
	headers   map[string]string // 自定义请求头
}

// Name 通知器名称
func (n *SlackNotifier) Name() string {
	return "slack"
}

// Init 初始化
func (n *SlackNotifier) Init(webhook, channel, username, iconEmoji string, customHeaders map[string]string) error {
	if webhook == "" {
		return fmt.Errorf("webhook is required")
	}

	n.webhook = webhook
	n.channel = channel
	n.username = username
	n.iconEmoji = iconEmoji
	n.headers = customHeaders

	return nil
}

// Send 发送通知
func (n *SlackNotifier) Send(event *base.Event) error {
	// 构建 Slack 消息（使用 Block Kit）
	msg := n.buildBlockKitMessage(event)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", n.webhook, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range n.headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack api error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// Close 关闭连接
func (n *SlackNotifier) Close() error {
	return nil
}

// buildBlockKitMessage 构建 Slack Block Kit 消息
func (n *SlackNotifier) buildBlockKitMessage(event *base.Event) map[string]interface{} {
	blocks := []map[string]interface{}{
		// Header
		{
			"type": "header",
			"text": map[string]interface{}{
				"type":  "plain_text",
				"text":  fmt.Sprintf("%s %s", n.getSeverityEmoji(event.Severity), event.Type),
				"emoji": true,
			},
		},
		// Divider
		{
			"type": "divider",
		},
		// Message
		{
			"type": "section",
			"fields": []map[string]interface{}{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Type:*\n%s", event.Type),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Severity:*\n%s", n.getSeverityText(event.Severity)),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Time:*\n%s", event.Timestamp.Format("2006-01-02 15:04:05")),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Host:*\n%s", event.Target),
				},
			},
		},
		// Message details
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*Message:*\n%s", event.Message),
			},
		},
	}

	// 添加详细数据
	if len(event.RawData) > 0 {
		detailsText := "*Details:*\n"
		for k, v := range event.RawData {
			detailsText += fmt.Sprintf("• %s: `%v`\n", k, v)
		}

		blocks = append(blocks, map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": detailsText,
			},
		})
	}

	// 添加 Footer
	blocks = append(blocks, map[string]interface{}{
		"type": "context",
		"elements": []map[string]interface{}{
			{
				"type": "mrkdwn",
				"text": "From Siriusec MCP Monitoring System",
			},
		},
	})

	message := map[string]interface{}{
		"blocks": blocks,
	}

	// 可选的覆盖设置
	if n.channel != "" {
		message["channel"] = n.channel
	}
	if n.username != "" {
		message["username"] = n.username
	}
	if n.iconEmoji != "" {
		message["icon_emoji"] = n.iconEmoji
	}

	return message
}

// getSeverityEmoji 获取严重程度对应的 Emoji
func (n *SlackNotifier) getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	default:
		return "📢"
	}
}

// getSeverityText 获取严重程度文本
func (n *SlackNotifier) getSeverityText(severity string) string {
	switch severity {
	case "critical":
		return "Critical"
	case "warning":
		return "Warning"
	case "info":
		return "Info"
	default:
		return "Unknown"
	}
}
