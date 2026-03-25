/*
WebAPI 通知器

将告警事件推送到任意 HTTP 端点
*/
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"siriusec-mcp/internal/plugins/base"
	"time"
)

// WebAPINotifier WebAPI 通知器
type WebAPINotifier struct {
	url     string            // HTTP 端点 URL
	method  string            // HTTP 方法，默认 POST
	timeout time.Duration     // 超时时间
	headers map[string]string // 自定义请求头
	client  *http.Client      // HTTP 客户端
}

// Name 通知器名称
func (n *WebAPINotifier) Name() string {
	return "webapi"
}

// Init 初始化
func (n *WebAPINotifier) Init(url string, customHeaders map[string]string) error {
	if url == "" {
		return fmt.Errorf("url is required")
	}

	n.url = url
	n.method = "POST"
	n.timeout = 30 * time.Second
	n.headers = customHeaders

	n.client = &http.Client{
		Timeout: n.timeout,
	}

	return nil
}

// Send 发送通知
func (n *WebAPINotifier) Send(event *base.Event) error {
	// 序列化事件为 JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 创建 HTTP 请求
	ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, n.method, n.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Close 关闭连接
func (n *WebAPINotifier) Close() error {
	n.client.CloseIdleConnections()
	return nil
}
