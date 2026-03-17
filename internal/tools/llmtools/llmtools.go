// Package llmtools 提供基于 LLM 的智能工具
package llmtools

import (
	"context"
	"fmt"
	"siriusec-mcp/internal/llm"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/mcp"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools 注册 LLM 工具
func RegisterTools(s *mcp.Server) {
	// 智能诊断分析工具
	s.RegisterTool("smart_diagnose",
		mcpp.NewTool("smart_diagnose",
			mcpp.WithDescription("使用 AI 智能分析系统问题，提供诊断建议（需要配置 DASHSCOPE_API_KEY）"),
			mcpp.WithString("symptom", mcpp.Required(), mcpp.Description("系统症状描述")),
			mcpp.WithString("context", mcpp.Description("相关上下文信息，如错误日志、系统配置等")),
		),
		handleSmartDiagnose,
		"sysom_llm",
	)

	// 日志分析工具
	s.RegisterTool("analyze_logs",
		mcpp.NewTool("analyze_logs",
			mcpp.WithDescription("使用 AI 分析系统日志，识别异常和潜在问题（需要配置 DASHSCOPE_API_KEY）"),
			mcpp.WithString("logs", mcpp.Required(), mcpp.Description("日志内容")),
			mcpp.WithString("log_type", mcpp.Description("日志类型，如 syslog, dmesg, application"), mcpp.DefaultString("syslog")),
		),
		handleAnalyzeLogs,
		"sysom_llm",
	)

	// 系统优化建议工具
	s.RegisterTool("optimization_advice",
		mcpp.NewTool("optimization_advice",
			mcpp.WithDescription("根据系统信息提供 AI 智能优化建议（需要配置 DASHSCOPE_API_KEY）"),
			mcpp.WithString("system_info", mcpp.Required(), mcpp.Description("系统信息，如 CPU、内存、磁盘使用情况")),
			mcpp.WithString("workload_type", mcpp.Description("工作负载类型，如 web, database, compute"), mcpp.DefaultString("general")),
		),
		handleOptimizationAdvice,
		"sysom_llm",
	)
}

// handleSmartDiagnose 处理智能诊断请求
func handleSmartDiagnose(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	symptom := request.GetString("symptom", "")
	context := request.GetString("context", "")

	if symptom == "" {
		return nil, fmt.Errorf("symptom is required")
	}

	client, err := llm.NewClient()
	if err != nil {
		logger.Sugar.Errorw("LLM client not available", "error", err)
		return nil, fmt.Errorf("LLM client not available, please configure DASHSCOPE_API_KEY: %w", err)
	}

	prompt := fmt.Sprintf(`你是一位资深的 Linux 系统运维专家。请根据以下症状和上下文分析问题原因，并提供解决方案。

症状: %s

上下文信息:
%s

请提供:
1. 可能的原因分析
2. 诊断步骤
3. 解决方案
4. 预防措施`, symptom, context)

	response, err := client.Complete(prompt)
	if err != nil {
		logger.Sugar.Errorw("LLM request failed", "error", err)
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	result := fmt.Sprintf("【AI 智能诊断分析】\n\n使用模型: %s\n\n%s", client.GetModel(), response)
	return mcpp.NewToolResultText(result), nil
}

// handleAnalyzeLogs 处理日志分析请求
func handleAnalyzeLogs(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	logs := request.GetString("logs", "")
	logType := request.GetString("log_type", "syslog")

	if logs == "" {
		return nil, fmt.Errorf("logs is required")
	}

	client, err := llm.NewClient()
	if err != nil {
		logger.Sugar.Errorw("LLM client not available", "error", err)
		return nil, fmt.Errorf("LLM client not available, please configure DASHSCOPE_API_KEY: %w", err)
	}

	prompt := fmt.Sprintf(`你是一位日志分析专家。请分析以下 %s 日志，识别异常和潜在问题。

日志内容:
%s

请提供:
1. 日志概览
2. 发现的异常和错误
3. 严重级别评估
4. 建议的处理措施`, logType, logs)

	response, err := client.Complete(prompt)
	if err != nil {
		logger.Sugar.Errorw("LLM request failed", "error", err)
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	result := fmt.Sprintf("【AI 日志分析】\n\n日志类型: %s\n使用模型: %s\n\n%s", logType, client.GetModel(), response)
	return mcpp.NewToolResultText(result), nil
}

// handleOptimizationAdvice 处理优化建议请求
func handleOptimizationAdvice(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	systemInfo := request.GetString("system_info", "")
	workloadType := request.GetString("workload_type", "general")

	if systemInfo == "" {
		return nil, fmt.Errorf("system_info is required")
	}

	client, err := llm.NewClient()
	if err != nil {
		logger.Sugar.Errorw("LLM client not available", "error", err)
		return nil, fmt.Errorf("LLM client not available, please configure DASHSCOPE_API_KEY: %w", err)
	}

	prompt := fmt.Sprintf(`你是一位 Linux 系统优化专家。请根据以下系统信息提供优化建议。

系统信息:
%s

工作负载类型: %s

请提供:
1. 当前系统瓶颈分析
2. 内核参数优化建议
3. 资源配置优化
4. 监控建议`, systemInfo, workloadType)

	response, err := client.Complete(prompt)
	if err != nil {
		logger.Sugar.Errorw("LLM request failed", "error", err)
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	result := fmt.Sprintf("【AI 优化建议】\n\n工作负载类型: %s\n使用模型: %s\n\n%s", workloadType, client.GetModel(), response)
	return mcpp.NewToolResultText(result), nil
}
