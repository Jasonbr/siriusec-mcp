package integration

import (
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/am"
	"siriusec-mcp/internal/tools/memdiag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server := mcp.NewServer("Test Server", "1.0.0")
	assert.NotNil(t, server)
	assert.NotNil(t, server.GetMCPServer())
}

func TestRegisterTools(t *testing.T) {
	server := mcp.NewServer("Test Server", "1.0.0")

	// 注册AM工具
	am.RegisterTools(server)

	// 验证工具已注册
	tool := server.GetTool("list_all_instances")
	assert.NotNil(t, tool)
	assert.Equal(t, "list_all_instances", tool.Name)
	assert.Contains(t, tool.Tags, "sysom_am")

	// 验证所有AM工具
	amTools := server.GetToolsByTag("sysom_am")
	assert.GreaterOrEqual(t, len(amTools), 4) // list_all_instances, list_instances, list_clusters, list_pods_of_instance
}

func TestRegisterMemDiagTools(t *testing.T) {
	server := mcp.NewServer("Test Server", "1.0.0")

	// 注册内存诊断工具
	memdiag.RegisterTools(server)

	// 验证工具已注册
	tool := server.GetTool("memgraph")
	assert.NotNil(t, tool)
	assert.Equal(t, "memgraph", tool.Name)
	assert.Contains(t, tool.Tags, "sysom_memdiag")

	// 验证所有内存诊断工具
	memTools := server.GetToolsByTag("sysom_memdiag")
	assert.GreaterOrEqual(t, len(memTools), 3) // memgraph, javamem, oomcheck
}

func TestGetAllTools(t *testing.T) {
	server := mcp.NewServer("Test Server", "1.0.0")

	// 注册所有工具
	am.RegisterTools(server)
	memdiag.RegisterTools(server)

	// 获取所有工具
	allTools := server.GetAllTools()
	assert.GreaterOrEqual(t, len(allTools), 7) // 4个AM工具 + 3个内存诊断工具
}
