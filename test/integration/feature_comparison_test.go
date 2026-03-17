/*
功能对比测试

对比Go版本和Python版本的功能覆盖
*/
package integration

import (
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/am"
	"siriusec-mcp/internal/tools/crashagent"
	"siriusec-mcp/internal/tools/initial"
	"siriusec-mcp/internal/tools/iodiag"
	"siriusec-mcp/internal/tools/memdiag"
	"siriusec-mcp/internal/tools/netdiag"
	"siriusec-mcp/internal/tools/otherdiag"
	"siriusec-mcp/internal/tools/scheddiag"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Python版本的功能清单 (实际有效的工具，metrics_mcp.py中的工具被注释掉了)
var pythonTools = map[string][]string{
	"am": {
		"list_all_instances",
		"list_instances",
		"list_clusters",
		"list_pods_of_instance",
	},
	"memdiag": {
		"memgraph",
		"javamem",
		"oomcheck",
	},
	"iodiag": {
		"iofsstat",
		"iodiagnose",
	},
	"netdiag": {
		"packetdrop",
		"netjitter",
	},
	"scheddiag": {
		"delay",
		"loadtask",
	},
	"otherdiag": {
		"vmcore_analysis",
		"disk_analysis",
	},
	"crashagent": {
		"list_vmcores",
		"get_vmcore_detail",
		"analyze_vmcore",
		"delete_vmcore",
	},
	"initial": {
		"check_sysom_initialed",
		"initial_sysom",
	},
}

// Python版本总工具数
const totalPythonTools = 21

// registerAllTools 注册所有工具
func registerAllTools(s *mcp.Server) {
	am.RegisterTools(s)
	memdiag.RegisterTools(s)
	iodiag.RegisterTools(s)
	netdiag.RegisterTools(s)
	scheddiag.RegisterTools(s)
	otherdiag.RegisterTools(s)
	crashagent.RegisterTools(s)
	initial.RegisterTools(s)
}

// TestAMToolsComplete 测试AM工具是否完整
func TestAMToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["am"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "AM工具 %s 应该已注册", toolName)
	}

	// 验证标签
	amTools := server.GetToolsByTag("sysom_am")
	assert.Equal(t, len(expectedTools), len(amTools), "AM工具数量应该匹配")
}

// TestMemDiagToolsComplete 测试内存诊断工具是否完整
func TestMemDiagToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["memdiag"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "内存诊断工具 %s 应该已注册", toolName)
	}

	// 验证标签
	memTools := server.GetToolsByTag("sysom_memdiag")
	assert.Equal(t, len(expectedTools), len(memTools), "内存诊断工具数量应该匹配")
}

// TestIODiagToolsComplete 测试IO诊断工具是否完整
func TestIODiagToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["iodiag"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "IO诊断工具 %s 应该已注册", toolName)
	}

	// 验证标签
	ioTools := server.GetToolsByTag("sysom_iodiag")
	assert.Equal(t, len(expectedTools), len(ioTools), "IO诊断工具数量应该匹配")
}

// TestNetDiagToolsComplete 测试网络诊断工具是否完整
func TestNetDiagToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["netdiag"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "网络诊断工具 %s 应该已注册", toolName)
	}

	// 验证标签
	netTools := server.GetToolsByTag("sysom_netdiagnose")
	assert.Equal(t, len(expectedTools), len(netTools), "网络诊断工具数量应该匹配")
}

// TestSchedDiagToolsComplete 测试调度诊断工具是否完整
func TestSchedDiagToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["scheddiag"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "调度诊断工具 %s 应该已注册", toolName)
	}

	// 验证标签
	schedTools := server.GetToolsByTag("sysom_scheddiag")
	assert.Equal(t, len(expectedTools), len(schedTools), "调度诊断工具数量应该匹配")
}

// TestOtherDiagToolsComplete 测试其他诊断工具是否完整
func TestOtherDiagToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["otherdiag"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "其他诊断工具 %s 应该已注册", toolName)
	}

	// 验证标签
	otherTools := server.GetToolsByTag("sysom_otherdiag")
	assert.Equal(t, len(expectedTools), len(otherTools), "其他诊断工具数量应该匹配")
}

// TestCrashAgentToolsComplete 测试崩溃诊断代理工具是否完整
func TestCrashAgentToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["crashagent"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "崩溃诊断代理工具 %s 应该已注册", toolName)
	}

	// 验证标签
	crashTools := server.GetToolsByTag("sysom_crash_agent")
	assert.Equal(t, len(expectedTools), len(crashTools), "崩溃诊断代理工具数量应该匹配")
}

// TestInitialToolsComplete 测试初始化服务工具是否完整
func TestInitialToolsComplete(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	expectedTools := pythonTools["initial"]
	for _, toolName := range expectedTools {
		tool := server.GetTool(toolName)
		assert.NotNil(t, tool, "初始化服务工具 %s 应该已注册", toolName)
	}

	// 验证标签
	initialTools := server.GetToolsByTag("sysom_initial")
	assert.Equal(t, len(expectedTools), len(initialTools), "初始化服务工具数量应该匹配")
}

// TestToolParameters 测试工具参数是否一致
func TestToolParameters(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	// 测试 list_all_instances 参数
	tool := server.GetTool("list_all_instances")
	assert.NotNil(t, tool)
	assert.NotNil(t, tool.Schema)
	assert.Equal(t, "list_all_instances", tool.Schema.Name)

	// 测试 memgraph 参数
	tool = server.GetTool("memgraph")
	assert.NotNil(t, tool)
	assert.NotNil(t, tool.Schema)
	assert.Equal(t, "memgraph", tool.Schema.Name)
}

// TestAllToolsImplemented 测试所有工具是否已实现
func TestAllToolsImplemented(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	// 检查所有工具是否都已实现
	allCategories := []string{"am", "memdiag", "iodiag", "netdiag", "scheddiag", "otherdiag", "crashagent", "initial"}

	for _, category := range allCategories {
		for _, toolName := range pythonTools[category] {
			tool := server.GetTool(toolName)
			assert.NotNil(t, tool, "工具 %s (分类: %s) 应该已注册", toolName, category)
		}
	}
}

// TestTotalToolCount 测试工具总数
func TestTotalToolCount(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	allTools := server.GetAllTools()

	// Python版本总共有21个工具(metrics_mcp.py中的工具被注释掉了)
	t.Logf("Go版本已实现工具数: %d", len(allTools))
	t.Logf("Python版本总工具数: %d", totalPythonTools)
	t.Logf("功能覆盖率: %.1f%%", float64(len(allTools))/float64(totalPythonTools)*100)

	// 断言实现了所有工具
	assert.Equal(t, totalPythonTools, len(allTools), "应该实现全部%d个工具", totalPythonTools)
}

// TestToolCategories 测试工具分类
func TestToolCategories(t *testing.T) {
	server := mcp.NewServer("Test", "1.0.0")
	registerAllTools(server)

	// 测试AM分类
	amTools := server.GetToolsByTag("sysom_am")
	assert.Len(t, amTools, 4, "AM分类应该有4个工具")

	// 测试内存诊断分类
	memTools := server.GetToolsByTag("sysom_memdiag")
	assert.Len(t, memTools, 3, "内存诊断分类应该有3个工具")

	// 测试IO诊断分类
	ioTools := server.GetToolsByTag("sysom_iodiag")
	assert.Len(t, ioTools, 2, "IO诊断分类应该有2个工具")

	// 测试网络诊断分类
	netTools := server.GetToolsByTag("sysom_netdiagnose")
	assert.Len(t, netTools, 2, "网络诊断分类应该有2个工具")

	// 测试调度诊断分类
	schedTools := server.GetToolsByTag("sysom_scheddiag")
	assert.Len(t, schedTools, 2, "调度诊断分类应该有2个工具")

	// 测试其他诊断分类
	otherTools := server.GetToolsByTag("sysom_otherdiag")
	assert.Len(t, otherTools, 2, "其他诊断分类应该有2个工具")

	// 测试崩溃诊断代理分类
	crashTools := server.GetToolsByTag("sysom_crash_agent")
	assert.Len(t, crashTools, 4, "崩溃诊断代理分类应该有4个工具")

	// 测试初始化服务分类
	initialTools := server.GetToolsByTag("sysom_initial")
	assert.Len(t, initialTools, 2, "初始化服务分类应该有2个工具")
}
