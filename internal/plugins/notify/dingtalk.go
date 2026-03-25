/*
钉钉通知器

通过钉钉机器人发送告警消息
*/
package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"siriusec-mcp/internal/plugins/base"
	"time"
)

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
	webhook string            // 钉钉机器人 webhook URL
	secret  string            // 加签密钥（可选）
	headers map[string]string // 自定义请求头
}

// Name 通知器名称
func (n *DingTalkNotifier) Name() string {
	return "dingtalk"
}

// Init 初始化
func (n *DingTalkNotifier) Init(webhook, secret string, customHeaders map[string]string) error {
	if webhook == "" {
		return fmt.Errorf("webhook is required")
	}

	n.webhook = webhook
	n.secret = secret
	n.headers = customHeaders

	return nil
}

// Send 发送通知
func (n *DingTalkNotifier) Send(event *base.Event) error {
	// 构建钉钉消息
	msg := n.buildMessage(event)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 生成带签名的 webhook URL
	sendURL := n.webhook
	if n.secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := n.generateSign(n.secret, timestamp)
		sendURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", n.webhook, timestamp, sign)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", sendURL, bytes.NewReader(data))
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

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("dingtalk api error: %v", result["errmsg"])
	}

	return nil
}

// Close 关闭连接
func (n *DingTalkNotifier) Close() error {
	return nil
}

// buildMessage 构建钉钉消息
func (n *DingTalkNotifier) buildMessage(event *base.Event) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": fmt.Sprintf("%s告警：%s", n.getSeverityText(event.Severity), event.Type),
			"text":  n.buildMarkdownMessage(event),
		},
		"at": map[string]interface{}{
			"isAtAll": true,
		},
	}
}

// buildMarkdownMessage 构建 Markdown 消息
func (n *DingTalkNotifier) buildMarkdownMessage(event *base.Event) string {
	severityEmoji := n.getSeverityEmoji(event.Severity)
	timeStr := event.Timestamp.Format("2006-01-02 15:04:05")

	text := fmt.Sprintf("## %s %s告警\n\n", severityEmoji, n.getSeverityText(event.Severity))
	text += fmt.Sprintf("**告警类型**: %s\n", event.Type)
	text += fmt.Sprintf("**严重程度**: %s\n", event.Severity)
	text += fmt.Sprintf("**告警时间**: %s\n\n", timeStr)
	text += fmt.Sprintf("### %s\n\n", event.Message)

	// 添加详细信息
	if len(event.RawData) > 0 {
		text += "**详细数据**:\n"
		for k, v := range event.RawData {
			text += fmt.Sprintf("- %s: %v\n", k, v)
		}
		text += "\n"
	}

	// 添加标签
	if len(event.Labels) > 0 {
		text += "**标签**:\n"
		for k, v := range event.Labels {
			text += fmt.Sprintf("- %s: %s\n", k, v)
		}
		text += "\n"
	}

	text += fmt.Sprintf("> 来自：Siriusec MCP 监控系统")

	return text
}

// getSeverityEmoji 获取严重程度对应的 Emoji
func (n *DingTalkNotifier) getSeverityEmoji(severity string) string {
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
func (n *DingTalkNotifier) getSeverityText(severity string) string {
	switch severity {
	case "critical":
		return "严重"
	case "warning":
		return "警告"
	case "info":
		return "提示"
	default:
		return "未知"
	}
}

// generateSign 生成钉钉签名
func (n *DingTalkNotifier) generateSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(signature)
}
