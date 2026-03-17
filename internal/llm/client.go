// Package llm 提供大语言模型客户端
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"siriusec-mcp/internal/config"
	"time"
)

const (
	// ProviderDashScope 阿里云 DashScope
	ProviderDashScope = "dashscope"
	// ProviderOpenAI OpenAI
	ProviderOpenAI = "openai"
	// ProviderAzure Azure OpenAI
	ProviderAzure = "azure"
	// ProviderCustom 自定义
	ProviderCustom = "custom"

	// DashScopeAPIEndpoint 阿里云 DashScope API 端点
	DashScopeAPIEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	// OpenAIAPIEndpoint OpenAI API 端点
	OpenAIAPIEndpoint = "https://api.openai.com/v1/chat/completions"
)

// Client LLM 客户端
type Client struct {
	provider    string
	apiKey      string
	endpoint    string
	model       string
	temperature float64
	maxTokens   int
	httpClient  *http.Client
}

// Message 对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request LLM 请求
type Request struct {
	Model  string `json:"model"`
	Input  Input  `json:"input"`
	Params Params `json:"parameters,omitempty"`
}

// Input 输入
type Input struct {
	Messages []Message `json:"messages"`
}

// Params 参数
type Params struct {
	ResultFormat string  `json:"result_format,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
}

// Response LLM 响应
type Response struct {
	Output    Output `json:"output"`
	Usage     Usage  `json:"usage"`
	RequestID string `json:"request_id"`
}

// Output 输出
type Output struct {
	Text    string   `json:"text"`
	Choices []Choice `json:"choices,omitempty"`
}

// Choice 选择
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 用量
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// NewClient 创建 LLM 客户端
func NewClient() (*Client, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// 获取 API Key（优先使用新配置）
	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		// 兼容旧配置
		apiKey = cfg.LLM.DashScopeAPIKey
	}
	if apiKey == "" {
		return nil, fmt.Errorf("LLM API Key not configured, please set LLM_API_KEY or DASHSCOPE_API_KEY")
	}

	// 确定提供商
	provider := cfg.LLM.Provider
	if provider == "" {
		provider = ProviderDashScope
	}

	// 确定端点
	endpoint := cfg.LLM.BaseURL
	if endpoint == "" {
		switch provider {
		case ProviderOpenAI:
			endpoint = OpenAIAPIEndpoint
		case ProviderDashScope:
			endpoint = DashScopeAPIEndpoint
		case ProviderAzure, ProviderCustom:
			// Azure 和 Custom 必须提供 BaseURL
			return nil, fmt.Errorf("LLM_BASE_URL is required for provider: %s", provider)
		default:
			endpoint = DashScopeAPIEndpoint
		}
	}

	// 确定模型
	model := cfg.LLM.Model
	if model == "" {
		switch provider {
		case ProviderOpenAI:
			model = "gpt-3.5-turbo"
		case ProviderDashScope:
			model = "qwen-turbo"
		default:
			model = "qwen-turbo"
		}
	}

	// 获取超时时间
	timeout := cfg.LLM.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	return &Client{
		provider:    provider,
		apiKey:      apiKey,
		endpoint:    endpoint,
		model:       model,
		temperature: cfg.LLM.Temperature,
		maxTokens:   cfg.LLM.MaxTokens,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}, nil
}

// NewClientWithModel 创建指定模型的 LLM 客户端
func NewClientWithModel(model string) (*Client, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	client.model = model
	return client, nil
}

// Chat 进行对话
func (c *Client) Chat(messages []Message) (*Response, error) {
	// 使用配置的参数，如果没有则使用默认值
	temperature := c.temperature
	if temperature == 0 {
		temperature = 0.7
	}
	maxTokens := c.maxTokens
	if maxTokens == 0 {
		maxTokens = 1500
	}

	req := Request{
		Model: c.model,
		Input: Input{
			Messages: messages,
		},
		Params: Params{
			ResultFormat: "message",
			Temperature:  temperature,
			MaxTokens:    maxTokens,
		},
	}

	return c.sendRequest(req)
}

// ChatWithSystem 带系统提示的对话
func (c *Client) ChatWithSystem(systemPrompt string, userMessage string) (*Response, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}
	return c.Chat(messages)
}

// Complete 文本补全
func (c *Client) Complete(prompt string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := c.Chat(messages)
	if err != nil {
		return "", err
	}

	if len(resp.Output.Choices) > 0 {
		return resp.Output.Choices[0].Message.Content, nil
	}

	return resp.Output.Text, nil
}

// sendRequest 发送请求
func (c *Client) sendRequest(req Request) (*Response, error) {
	// 根据提供商构建不同的请求格式
	var jsonData []byte
	var err error

	switch c.provider {
	case ProviderOpenAI, ProviderAzure, ProviderCustom:
		// OpenAI 兼容格式
		openAIReq := c.toOpenAIRequest(req)
		jsonData, err = json.Marshal(openAIReq)
	default:
		// DashScope 格式
		jsonData, err = json.Marshal(req)
	}

	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", httpResp.StatusCode, string(body))
	}

	// 根据提供商解析不同的响应格式
	switch c.provider {
	case ProviderOpenAI, ProviderAzure, ProviderCustom:
		return c.parseOpenAIResponse(body)
	default:
		var resp Response
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("unmarshal response failed: %w", err)
		}
		return &resp, nil
	}
}

// OpenAIRequest OpenAI 兼容请求格式
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// OpenAIResponse OpenAI 兼容响应格式
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int     `json:"index"`
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// toOpenAIRequest 转换为 OpenAI 请求格式
func (c *Client) toOpenAIRequest(req Request) OpenAIRequest {
	return OpenAIRequest{
		Model:       req.Model,
		Messages:    req.Input.Messages,
		Temperature: req.Params.Temperature,
		MaxTokens:   req.Params.MaxTokens,
		Stream:      false,
	}
}

// parseOpenAIResponse 解析 OpenAI 响应
func (c *Client) parseOpenAIResponse(body []byte) (*Response, error) {
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("unmarshal OpenAI response failed: %w", err)
	}

	// 转换为统一格式
	resp := &Response{
		RequestID: openAIResp.ID,
		Usage: Usage{
			InputTokens:  openAIResp.Usage.PromptTokens,
			OutputTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:  openAIResp.Usage.TotalTokens,
		},
	}

	if len(openAIResp.Choices) > 0 {
		resp.Output.Choices = []Choice{
			{
				Message: openAIResp.Choices[0].Message,
			},
		}
		resp.Output.Text = openAIResp.Choices[0].Message.Content
	}

	return resp, nil
}

// SetModel 设置模型
func (c *Client) SetModel(model string) {
	c.model = model
}

// GetModel 获取当前模型
func (c *Client) GetModel() string {
	return c.model
}

// SupportedModels 支持的模型列表
func SupportedModels() []string {
	return []string{
		"qwen-turbo",           // 通义千问 Turbo
		"qwen-plus",            // 通义千问 Plus
		"qwen-max",             // 通义千问 Max
		"qwen-max-longcontext", // 通义千问 Max 长上下文
	}
}
