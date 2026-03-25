package httpcheck
/*
HTTP 可用性检测插件

监控 HTTP/HTTPS 服务的可用性、响应时间、SSL 证书等
*/
package httpcheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"siriusec-mcp/internal/plugins/base"
)

// HTTPCheckPlugin HTTP 检查插件
type HTTPCheckPlugin struct {
	config         base.PluginConfig
	status         base.PluginStatus
	url            string        // 目标 URL
	method         string        // HTTP 方法
	timeout        time.Duration // 超时时间
	expectCode     int           // 期望的状态码
	expectContent  string        // 期望包含的内容（可选）
	responseThreshold int        // 响应时间阈值（毫秒）
	certExpireThreshold int      // 证书过期阈值（天）
}

// Name 插件名称
func (p *HTTPCheckPlugin) Name() string {
	return "httpcheck"
}

// Description 插件描述
func (p *HTTPCheckPlugin) Description() string {
	return "HTTP/HTTPS服务可用性、响应时间、SSL 证书监控"
}

// Init 初始化插件
func (p *HTTPCheckPlugin) Init(config base.PluginConfig) error {
	p.config = config
	p.status = base.StatusStopped

	// 从配置中获取参数
	if url, ok := config.Params["url"].(string); ok {
		p.url = url
	} else {
		p.url = "http://localhost"
	}

	if method, ok := config.Params["method"].(string); ok {
		p.method = method
	} else {
		p.method = "GET"
	}

	if timeout, ok := config.Params["timeout"].(float64); ok {
		p.timeout = time.Duration(timeout) * time.Second
	} else {
		p.timeout = 30 * time.Second
	}

	if code, ok := config.Params["expect_code"].(float64); ok {
		p.expectCode = int(code)
	} else {
		p.expectCode = 200
	}

	if content, ok := config.Params["expect_content"].(string); ok {
		p.expectContent = content
	}

	if threshold, ok := config.Params["response_threshold"].(float64); ok {
		p.responseThreshold = int(threshold)
	} else {
		p.responseThreshold = 1000 // 1 秒
	}

	if threshold, ok := config.Params["cert_expire_threshold"].(float64); ok {
		p.certExpireThreshold = int(threshold)
	} else {
		p.certExpireThreshold = 30 // 30 天
	}

	return nil
}

// Start 启动监控循环
func (p *HTTPCheckPlugin) Start(ctx context.Context) error {
	p.status = base.StatusRunning

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			event, err := p.Check(ctx)
			if err != nil {
				return fmt.Errorf("httpcheck failed: %w", err)
			}

			if event != nil {
				fmt.Printf("[HTTPCHECK Alert] %s: %s\n", event.Type, event.Message)
			}

		case <-ctx.Done():
			p.status = base.StatusStopped
			return nil
		}
	}
}

// Stop 停止监控
func (p *HTTPCheckPlugin) Stop() error {
	p.status = base.StatusStopped
	return nil
}

// Status 获取插件状态
func (p *HTTPCheckPlugin) Status() base.PluginStatus {
	return p.status
}

// Check 执行一次检查
func (p *HTTPCheckPlugin) Check(ctx context.Context) (*base.Event, error) {
	startTime := time.Now()

	// 创建 HTTP 请求
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, p.method, p.url, nil)
	if err != nil {
		return &base.Event{
			ID:        fmt.Sprintf("http_error_%d", time.Now().Unix()),
			Type:      "http_request_error",
			Severity:  "critical",
			Timestamp: time.Now(),
			Target:    p.url,
			Metric:    "request_status",
			Value:     "error",
			Threshold: "success",
			Message:   fmt.Sprintf("HTTP 请求创建失败：%v", err),
			Labels: map[string]string{
				"plugin":   "httpcheck",
				"hostname": getHostname(),
				"severity": "critical",
				"url":      p.url,
			},
			RawData: map[string]interface{}{
				"error": err.Error(),
			},
		}, nil
	}

	// 禁用自动重定向
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 如果是 HTTPS，检查 SSL 证书
	if strings.HasPrefix(p.url, "https://") {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = transport
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return &base.Event{
			ID:        fmt.Sprintf("http_down_%d", time.Now().Unix()),
			Type:      "http_service_down",
			Severity:  "critical",
			Timestamp: time.Now(),
			Target:    p.url,
			Metric:    "service_status",
			Value:     "down",
			Threshold: "up",
			Message:   fmt.Sprintf("HTTP 服务不可用：%v", err),
			Labels: map[string]string{
				"plugin":   "httpcheck",
				"hostname": getHostname(),
				"severity": "critical",
				"url":      p.url,
			},
			RawData: map[string]interface{}{
				"error": err.Error(),
			},
		}, nil
	}
	defer resp.Body.Close()

	duration := time.Since(startTime).Milliseconds()

	// 检查状态码
	if resp.StatusCode != p.expectCode {
		return &base.Event{
			ID:        fmt.Sprintf("http_status_%d", time.Now().Unix()),
			Type:      "http_status_unexpected",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.url,
			Metric:    "status_code",
			Value:     resp.StatusCode,
			Threshold: float64(p.expectCode),
			Message:   fmt.Sprintf("HTTP 状态码异常：%d (期望：%d)", resp.StatusCode, p.expectCode),
			Labels: map[string]string{
				"plugin":   "httpcheck",
				"hostname": getHostname(),
				"severity": "warning",
				"url":      p.url,
			},
			RawData: map[string]interface{}{
				"status_code":   resp.StatusCode,
				"expect_code":   p.expectCode,
				"response_time": duration,
			},
		}, nil
	}

	// 检查响应时间
	if int(duration) > p.responseThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("http_slow_%d", time.Now().Unix()),
			Type:      "http_response_slow",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.url,
			Metric:    "response_time_ms",
			Value:     duration,
			Threshold: float64(p.responseThreshold),
			Message:   fmt.Sprintf("HTTP 响应过慢：%dms (阈值：%dms)", duration, p.responseThreshold),
			Labels: map[string]string{
				"plugin":   "httpcheck",
				"hostname": getHostname(),
				"severity": "warning",
				"url":      p.url,
			},
			RawData: map[string]interface{}{
				"response_time": duration,
				"threshold":     p.responseThreshold,
				"status_code":   resp.StatusCode,
			},
		}, nil
	}

	// 检查响应内容
	if p.expectContent != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil
		}

		if !strings.Contains(string(body), p.expectContent) {
			return &base.Event{
				ID:        fmt.Sprintf("http_content_%d", time.Now().Unix()),
				Type:      "http_content_mismatch",
				Severity:  "warning",
				Timestamp: time.Now(),
				Target:    p.url,
				Metric:    "content_check",
				Value:     "mismatch",
				Threshold: "match",
				Message:   fmt.Sprintf("HTTP 响应内容不匹配：未找到 \"%s\"", p.expectContent),
				Labels: map[string]string{
					"plugin":   "httpcheck",
					"hostname": getHostname(),
					"severity": "warning",
					"url":      p.url,
				},
				RawData: map[string]interface{}{
					"expect_content": p.expectContent,
					"body_length":    len(body),
					"status_code":    resp.StatusCode,
				},
			}, nil
		}
	}

	// 检查 SSL 证书有效期
	if strings.HasPrefix(p.url, "https://") {
		if certEvent := p.checkSSLCert(resp); certEvent != nil {
			return certEvent, nil
		}
	}

	return nil, nil
}

// checkSSLCert 检查 SSL 证书
func (p *HTTPCheckPlugin) checkSSLCert(resp *http.Response) *base.Event {
	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return nil
	}

	cert := resp.TLS.PeerCertificates[0]
	notAfter := cert.NotAfter
	daysUntilExpiry := int(time.Until(notAfter).Hours() / 24)

	if daysUntilExpiry < p.certExpireThreshold {
		return &base.Event{
			ID:        fmt.Sprintf("ssl_expire_%d", time.Now().Unix()),
			Type:      "ssl_certificate_expiring",
			Severity:  "warning",
			Timestamp: time.Now(),
			Target:    p.url,
			Metric:    "ssl_days_until_expiry",
			Value:     daysUntilExpiry,
			Threshold: float64(p.certExpireThreshold),
			Message:   fmt.Sprintf("SSL 证书即将过期：剩余 %d 天", daysUntilExpiry),
			Labels: map[string]string{
				"plugin":   "httpcheck",
				"hostname": getHostname(),
				"severity": "warning",
				"url":      p.url,
			},
			RawData: map[string]interface{}{
				"days_until_expiry": daysUntilExpiry,
				"expiry_date":       notAfter.Format("2006-01-02"),
				"issuer":            cert.Issuer.CommonName,
				"subject":           cert.Subject.CommonName,
			},
		}
	}

	return nil
}

// getHostname 获取主机名
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}
