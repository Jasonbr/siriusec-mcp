/*
腾讯云监控客户端

用于接入腾讯云 Cloud Monitor API
*/
package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// TencentClient 腾讯云客户端
type TencentClient struct {
	secretID   string
	secretKey  string
	region     string
	httpClient *http.Client
}

// NewTencentClient 创建腾讯云客户端
func NewTencentClient() (*TencentClient, error) {
	secretID := os.Getenv("TENCENT_SECRET_ID")
	secretKey := os.Getenv("TENCENT_SECRET_KEY")
	region := os.Getenv("TENCENT_REGION")

	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("Tencent credentials not configured")
	}

	if region == "" {
		region = "ap-guangzhou" // 默认广州区域
	}

	return &TencentClient{
		secretID:  secretID,
		secretKey: secretKey,
		region:    region,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetMonitorData 获取监控数据
func (c *TencentClient) GetMonitorData(instanceID string, metricName string, period int64) (*MonitorData, error) {
	// 构造请求参数
	params := map[string]interface{}{
		"Namespace":  "QCE/CVM", // 云服务器命名空间
		"MetricName": metricName,
		"Period":     period,
		"Instances": []map[string]string{
			{"InstanceId": instanceID},
		},
		"StartTime": time.Now().Add(-5 * time.Minute).Unix(),
		"EndTime":   time.Now().Unix(),
	}

	// 发送请求
	respBody, err := c.sendRequest("GetMonitorData", params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result MonitorDataResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.Response.Data, nil
}

// DescribeInstances 查询实例列表
func (c *TencentClient) DescribeInstances(zone string) ([]InstanceInfo, error) {
	params := map[string]interface{}{
		"Zone": zone,
	}

	respBody, err := c.sendRequest("DescribeInstances", params)
	if err != nil {
		return nil, err
	}

	var result InstancesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result.Response.InstanceSet, nil
}

// sendRequest 发送腾讯云 API 请求
func (c *TencentClient) sendRequest(action string, params map[string]interface{}) ([]byte, error) {
	// 准备请求体
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// 构造签名
	timestamp := time.Now().Unix()
	signature := c.sign(fmt.Sprintf("%d", timestamp), action, string(requestBody))

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", "https://monitor.tencentcloudapi.com", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Version", "2018-07-24")
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Region", c.region)
	req.Header.Set("Authorization", signature)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody := make([]byte, 4096)
	n, err := resp.Body.Read(respBody)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return respBody[:n], nil
}

// sign 生成腾讯云签名
func (c *TencentClient) sign(timestamp string, action string, payload string) string {
	secretDate := hmacSHA256(timestamp, "TC3"+c.secretKey)
	secretService := hmacSHA256("monitor", secretDate)
	secretSigning := hmacSHA256("tc3_request", secretService)

	canonicalRequest := fmt.Sprintf("POST\n/\n\ncontent-type:application/json\nhost:monitor.tencentcloudapi.com\nx-tc-action:%s\nx-tc-region:%s\nx-tc-timestamp:%s\nx-tc-version:2018-07-24\n\ncontent-type;host;x-tc-action;x-tc-region;x-tc-timestamp;x-tc-version\n%s",
		action, c.region, timestamp, sha256Sum(payload))

	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%s\n%s", timestamp, sha256Sum(canonicalRequest))
	signature := hmacSHA256(stringToSign, secretSigning)

	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s/monitor/tc3_request, SignedHeaders=content-type;host;x-tc-action;x-tc-region;x-tc-timestamp;x-tc-version, Signature=%s",
		c.secretID, timestamp, signature)
}

func hmacSHA256(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return string(h.Sum(nil))
}

func sha256Sum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// MonitorData 监控数据结构
type MonitorData struct {
	MetricName string      `json:"MetricName"`
	DataPoints []DataPoint `json:"DataPoints"`
}

// DataPoint 数据点
type DataPoint struct {
	Timestamp int64   `json:"Timestamp"`
	Value     float64 `json:"Value"`
}

// InstanceInfo 实例信息
type InstanceInfo struct {
	InstanceId   string `json:"InstanceId"`
	InstanceName string `json:"InstanceName"`
	State        string `json:"State"`
	Zone         string `json:"Zone"`
}

// MonitorDataResponse 监控数据响应
type MonitorDataResponse struct {
	Response struct {
		Data MonitorData `json:"Data"`
	} `json:"Response"`
}

// InstancesResponse 实例列表响应
type InstancesResponse struct {
	Response struct {
		InstanceSet []InstanceInfo `json:"InstanceSet"`
	} `json:"Response"`
}
