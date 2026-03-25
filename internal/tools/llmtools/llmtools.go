// Package llmtools 提供基于 LLM 的智能工具，支持自动调用诊断工具
package llmtools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"siriusec-mcp/internal/client"
	"siriusec-mcp/internal/llm"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/mcp"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	mcpp "github.com/mark3labs/mcp-go/mcp"
)

// DiagnosisPlan AI 生成的诊断计划
type DiagnosisPlan struct {
	Tools  []string `json:"tools"`
	Reason string   `json:"reason"`
}

// callDiagnosticTool 调用底层诊断工具（C 方案：gopsutil + shell 降级）
func callDiagnosticTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	logger.Sugar.Infow("Collecting system metrics", "tool", toolName)

	// 检查是否配置了云服务商
	cloudProvider := os.Getenv("CLOUD_PROVIDER")
	if cloudProvider != "" && params["instance_id"] != "" {
		// 使用云服务 API
		return collectCloudMetrics(ctx, cloudProvider, params)
	}

	// 默认：本地系统信息采集
	// 支持多种工具名称映射到相同的采集逻辑
	switch toolName {
	case "sched_diag", "cpu_diag", "load_task", "process_diag":
		// CPU/调度/进程/负载相关诊断
		return collectCPUMetrics(ctx)
	case "mem_diag", "memory_diag", "oom_diag":
		// 内存相关诊断
		return collectMemMetrics(ctx)
	case "io_diag", "disk_diag", "storage_diag":
		// IO/磁盘/存储相关诊断
		return collectIOMetrics(ctx)
	case "net_diag", "network_diag", "packet_diag":
		// 网络相关诊断
		return collectNetMetrics(ctx)
	default:
		// 对于未知工具，尝试根据关键词匹配
		lowerName := strings.ToLower(toolName)
		if strings.Contains(lowerName, "cpu") || strings.Contains(lowerName, "sched") || strings.Contains(lowerName, "load") || strings.Contains(lowerName, "process") {
			return collectCPUMetrics(ctx)
		}
		if strings.Contains(lowerName, "mem") || strings.Contains(lowerName, "oom") {
			return collectMemMetrics(ctx)
		}
		if strings.Contains(lowerName, "io") || strings.Contains(lowerName, "disk") {
			return collectIOMetrics(ctx)
		}
		if strings.Contains(lowerName, "net") || strings.Contains(lowerName, "packet") {
			return collectNetMetrics(ctx)
		}
		logger.Sugar.Warnw("Unknown diagnostic tool, using CPU metrics as fallback", "tool", toolName)
		return collectCPUMetrics(ctx)
	}
}

// collectCPUMetrics 收集 CPU 指标（C 方案：gopsutil 优先，shell 降级）
func collectCPUMetrics(ctx context.Context) (map[string]interface{}, error) {
	// 尝试方法 1: gopsutil
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		loadAvg, _ := load.Avg()
		processes, _ := process.Processes()

		var topProcs []string
		for i, p := range processes {
			if i >= 10 {
				break
			}
			name, _ := p.Name()
			cpuP, _ := p.CPUPercent()
			topProcs = append(topProcs, fmt.Sprintf("%s (PID:%d) - %.1f%% CPU", name, p.Pid, cpuP))
		}

		return map[string]interface{}{
			"cpu_usage":     fmt.Sprintf("%.1f%%", cpuPercent[0]),
			"load_average":  fmt.Sprintf("%.2f, %.2f, %.2f", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15),
			"top_processes": strings.Join(topProcs[:min(5, len(topProcs))], "\n"),
			"source":        "gopsutil",
		}, nil
	}

	// 降级方法 2: Shell 命令
	logger.Sugar.Warn("gopsutil failed, falling back to shell commands")
	uptime, _ := exec.CommandContext(ctx, "uptime").Output()
	top, _ := exec.CommandContext(ctx, "top", "-bn1", "-n", "1").Output()
	ps, _ := exec.CommandContext(ctx, "ps", "aux", "--sort=-%cpu").Output()

	lines := strings.Split(string(ps), "\n")
	var topProcs []string
	for i, line := range lines {
		if i > 0 && i <= 6 {
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				topProcs = append(topProcs, fmt.Sprintf("%s (PID:%s) - %s%% CPU", fields[10], fields[1], fields[2]))
			}
		}
	}

	return map[string]interface{}{
		"uptime":        strings.TrimSpace(string(uptime)),
		"top_output":    string(top),
		"top_processes": strings.Join(topProcs, "\n"),
		"source":        "shell",
	}, nil
}

// collectMemMetrics 收集内存指标
func collectMemMetrics(ctx context.Context) (map[string]interface{}, error) {
	// 尝试方法 1: gopsutil
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		swapInfo, _ := mem.SwapMemory()
		return map[string]interface{}{
			"memory_total": fmt.Sprintf("%.1f GB", float64(memInfo.Total)/1024/1024/1024),
			"memory_used":  fmt.Sprintf("%.1f GB (%.1f%%)", float64(memInfo.Used)/1024/1024/1024, memInfo.UsedPercent),
			"memory_free":  fmt.Sprintf("%.1f GB", float64(memInfo.Free)/1024/1024/1024),
			"swap_used":    fmt.Sprintf("%.1f GB (%.1f%%)", float64(swapInfo.Used)/1024/1024/1024, swapInfo.UsedPercent),
			"source":       "gopsutil",
		}, nil
	}

	// 降级方法 2: Shell 命令
	logger.Sugar.Warn("gopsutil failed, falling back to shell commands")
	free, _ := exec.CommandContext(ctx, "free", "-h").Output()
	lines := strings.Split(string(free), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 7 {
			return map[string]interface{}{
				"memory_info": string(free),
				"total":       fields[1],
				"used":        fields[2],
				"free":        fields[3],
				"source":      "shell",
			}, nil
		}
	}

	return map[string]interface{}{"error": "failed to parse memory info"}, nil
}

// collectIOMetrics 收集 IO 指标
func collectIOMetrics(ctx context.Context) (map[string]interface{}, error) {
	// TODO: 需要导入 disk 包
	// 暂时使用 shell 命令
	iostat, _ := exec.CommandContext(ctx, "iostat", "-x", "1", "1").Output()
	return map[string]interface{}{
		"iostat_output": string(iostat),
		"source":        "shell",
	}, nil
}

// collectNetMetrics 收集网络指标
func collectNetMetrics(ctx context.Context) (map[string]interface{}, error) {
	// TODO: 需要导入 net 包
	// 暂时使用 shell 命令
	netstat, _ := exec.CommandContext(ctx, "netstat", "-i").Output()
	return map[string]interface{}{
		"netstat_output": string(netstat),
		"source":         "shell",
	}, nil
}

// collectCloudMetrics 收集云服务商监控数据
func collectCloudMetrics(ctx context.Context, provider string, params map[string]interface{}) (map[string]interface{}, error) {
	instanceID := params["instance_id"].(string)

	switch provider {
	case "aliyun":
		// 阿里云 SysOM（后续集成）
		return collectAliyunMetrics(ctx, instanceID, params)
	case "tencent":
		// 腾讯云 Cloud Monitor
		return collectTencentMetrics(ctx, instanceID, params)
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// collectAliyunMetrics 阿里云监控数据采集
func collectAliyunMetrics(ctx context.Context, instanceID string, params map[string]interface{}) (map[string]interface{}, error) {
	// TODO: 集成阿里云 SysOM API
	return map[string]interface{}{
		"message":     "阿里云监控数据接入中...",
		"instance_id": instanceID,
		"provider":    "aliyun",
	}, nil
}

// collectTencentMetrics 腾讯云监控数据采集
func collectTencentMetrics(ctx context.Context, instanceID string, params map[string]interface{}) (map[string]interface{}, error) {
	// 创建腾讯云客户端
	tencentClient, err := client.NewTencentClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create tencent client: %w", err)
	}

	// 获取 CPU 使用率
	cpuData, err := tencentClient.GetMonitorData(instanceID, "CPUUtilization", 60)
	if err != nil {
		logger.Sugar.Warnw("Failed to get CPU data from Tencent", "error", err)
		return nil, err
	}

	// 获取内存使用率
	memData, err := tencentClient.GetMonitorData(instanceID, "MemoryUtilization", 60)
	if err != nil {
		logger.Sugar.Warnw("Failed to get Memory data from Tencent", "error", err)
		return nil, err
	}

	var cpuValue, memValue float64
	if len(cpuData.DataPoints) > 0 {
		cpuValue = cpuData.DataPoints[0].Value
	}
	if len(memData.DataPoints) > 0 {
		memValue = memData.DataPoints[0].Value
	}

	return map[string]interface{}{
		"cpu_usage":    fmt.Sprintf("%.1f%%", cpuValue),
		"memory_usage": fmt.Sprintf("%.1f%%", memValue),
		"instance_id":  instanceID,
		"provider":     "tencent",
		"source":       "tencent_cloud_api",
	}, nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractJSON 从响应中提取 JSON
func extractJSON(response string) string {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end < start {
		return response
	}
	return response[start : end+1]
}

// formatToolResults 格式化工具结果
func formatToolResults(results map[string]interface{}) string {
	var sb strings.Builder
	for toolName, result := range results {
		sb.WriteString(fmt.Sprintf("**%s**: %v\n", toolName, result))
	}
	return sb.String()
}

// handleSmartDiagnose 处理智能诊断请求 - 自动调用工具版本
func handleSmartDiagnose(ctx context.Context, request mcpp.CallToolRequest) (*mcpp.CallToolResult, error) {
	symptom := request.GetString("symptom", "")
	contextInfo := request.GetString("context", "")

	if symptom == "" {
		return nil, fmt.Errorf("symptom is required")
	}

	client, err := llm.NewClient()
	if err != nil {
		logger.Sugar.Errorw("LLM client not available", "error", err)
		return nil, fmt.Errorf("LLM client not available: %w", err)
	}

	// 第一步：AI 分析需要调用哪些诊断工具
	planningPrompt := fmt.Sprintf(`你是智能诊断调度器。根据症状决定需要调用哪些诊断工具获取实时数据。

症状：%s
上下文：%s

可用工具：
- sched_diag: CPU/进程调度诊断
- mem_diag: 内存诊断  
- io_diag: IO/存储诊断
- net_diag: 网络诊断
- load_task: 系统负载分析

返回 JSON: {"tools": ["工具列表"], "reason": "原因"}

示例：{"tools": ["sched_diag", "mem_diag"], "reason": "CPU 高可能与进程和内存有关"}`, symptom, contextInfo)

	planningResponse, err := client.Complete(planningPrompt)
	if err != nil {
		logger.Sugar.Errorw("AI planning failed", "error", err)
		return nil, fmt.Errorf("AI planning failed: %w", err)
	}

	// 第二步：解析 JSON，提取工具列表
	var plan DiagnosisPlan
	err = json.Unmarshal([]byte(extractJSON(planningResponse)), &plan)
	if err != nil {
		logger.Sugar.Warnw("Failed to parse diagnosis plan", "error", err, "response", planningResponse)
		// 如果解析失败，使用默认诊断计划
		plan.Tools = []string{"sched_diag", "mem_diag"}
		plan.Reason = "默认诊断：CPU 和内存是常见问题"
	}

	// 第三步：调用底层诊断工具，收集真实数据
	toolResults := make(map[string]interface{})
	for _, toolName := range plan.Tools {
		result, err := callDiagnosticTool(ctx, toolName, map[string]interface{}{
			"symptom": symptom,
			"context": contextInfo,
		})
		if err != nil {
			logger.Sugar.Errorw("Tool call failed", "tool", toolName, "error", err)
			toolResults[toolName] = map[string]string{"error": err.Error()}
			continue
		}
		toolResults[toolName] = result
	}

	// 第四步：AI 基于真实数据分析
	analysisPrompt := fmt.Sprintf(`你是资深 Linux 运维专家。请根据以下症状和实际诊断数据进行专业分析。

【症状描述】
症状：%s
上下文：%s

【实际诊断数据】
%s

【输出要求】
请按以下格式输出诊断报告：

## 根本原因分析

| 排名 | 原因 | 概率 | 依据 |
|:---|:---|:---|:---|
| 1 | 主要原因 | XX%% | 数据支撑 |
| 2 | 次要原因 | XX%% | 数据支撑 |

## 快速诊断步骤

1. **步骤名称**（预计X秒）
   `+"```"+`bash
   具体命令
   `+"```"+`

2. **步骤名称**（预计X秒）
   `+"```"+`bash
   具体命令
   `+"```"+`

## 解决方案

**紧急措施**：
- 立即执行的操作

**中期优化**：
- 代码或配置优化

**长期规划**：
- 架构或监控改进

## 关键结论

用1-2句话总结核心发现和行动建议。

**注意**：
- 使用标准 Markdown 表格，不要用 ASCII 图表
- 代码块使用 bash 标记
- 总字数控制在 500 字以内
- 避免使用特殊 Unicode 字符`, symptom, contextInfo, formatToolResults(toolResults))

	finalAnalysis, err := client.Complete(analysisPrompt)
	if err != nil {
		logger.Sugar.Errorw("AI analysis failed", "error", err)
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	resultText := fmt.Sprintf(`【智能诊断报告】

🤖 诊断模型：%s
📋 调用的工具：%v
⏱️ 数据来源：%s

---

%s`, client.GetModel(), plan.Tools, getDataSourceDescription(toolResults), finalAnalysis)

	return mcpp.NewToolResultText(resultText), nil
}

// getDataSourceDescription 描述数据来源
func getDataSourceDescription(results map[string]interface{}) string {
	hasError := false
	hasSuccess := false
	source := ""

	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if src, exists := resultMap["source"].(string); exists {
				source = src
			}
		}
		if errMap, ok := result.(map[string]string); ok && errMap["error"] != "" {
			hasError = true
		} else {
			hasSuccess = true
		}
	}

	if hasSuccess && !hasError {
		if source == "tencent_cloud_api" {
			return "腾讯云监控实时数据"
		} else if source == "gopsutil" {
			return "本机系统实时数据（gopsutil）"
		} else {
			return "本机系统实时数据（shell）"
		}
	} else if hasSuccess && hasError {
		return "部分实时数据 + 部分通用知识"
	} else {
		return "通用知识库（工具调用失败）"
	}
}

// RegisterTools 注册智能诊断工具
func RegisterTools(s *mcp.Server) {
	s.RegisterTool("smart_diagnose",
		mcpp.NewTool("smart_diagnose",
			mcpp.WithDescription("智能诊断工具，自动调用底层诊断工具并基于 AI 分析给出专业报告"),
			mcpp.WithString("symptom", mcpp.Required(), mcpp.Description("症状描述，如'CPU 使用率 100%'")),
			mcpp.WithString("context", mcpp.Description("上下文信息，如实例 ID、应用类型等")),
		),
		handleSmartDiagnose,
		"sysom_llmtools",
	)
}
