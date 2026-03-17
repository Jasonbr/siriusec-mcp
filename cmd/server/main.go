// Package main Siriusec MCP 服务器
// 统一的 MCP 服务器入口，聚合所有 MCP 服务
// 支持 stdio、SSE、streamable-http 三种运行模式
package main

import (
	"flag"
	"fmt"
	"os"
	"siriusec-mcp/internal/cli"
	"siriusec-mcp/internal/logger"
	"siriusec-mcp/internal/mcp"
	"siriusec-mcp/internal/tools/am"
	"siriusec-mcp/internal/tools/crashagent"
	"siriusec-mcp/internal/tools/initial"
	"siriusec-mcp/internal/tools/iodiag"
	"siriusec-mcp/internal/tools/llmtools"
	"siriusec-mcp/internal/tools/memdiag"
	"siriusec-mcp/internal/tools/netdiag"
	"siriusec-mcp/internal/tools/otherdiag"
	"siriusec-mcp/internal/tools/scheddiag"
	"siriusec-mcp/internal/version"

	"go.uber.org/zap"
)

var (
	// 版本信息，由构建脚本注入
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func init() {
	// 设置版本信息
	version.Version = Version
	version.GitCommit = GitCommit
	version.BuildTime = BuildTime
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		// 默认运行服务器
		runServer([]string{})
		return
	}

	switch args[0] {
	case "run":
		runServer(args[1:])
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printUsage()
	case "tools":
		handleToolsCommand(args[1:])
	case "config":
		handleConfigCommand(args[1:])
	case "selftest":
		handleSelfTestCommand(args[1:])
	case "mcptest":
		handleMCPTestCommand(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func runServer(args []string) {
	// 子命令参数
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	sse := fs.Bool("sse", false, "Run in SSE mode")
	streamableHTTP := fs.Bool("streamable-http", false, "Run in streamable-http mode")
	host := fs.String("host", "127.0.0.1", "Host to bind to (for SSE/streamable-http mode)")
	port := fs.Int("port", 7140, "Port to bind to (for SSE/streamable-http mode)")
	path := fs.String("path", "/mcp/unified", "Path for SSE/streamable-http endpoint")
	fs.Parse(args)

	// 创建 MCP 服务器
	server := mcp.NewServer("Siriusec MCP Server", version.Short())

	// 注册所有工具
	logger.Info("Registering tools...")
	am.RegisterTools(server)
	memdiag.RegisterTools(server)
	iodiag.RegisterTools(server)
	netdiag.RegisterTools(server)
	scheddiag.RegisterTools(server)
	otherdiag.RegisterTools(server)
	crashagent.RegisterTools(server)
	initial.RegisterTools(server)
	llmtools.RegisterTools(server)

	logger.Info("Tools loaded",
		zap.Int("count", len(server.GetAllTools())),
		zap.String("version", version.Short()))

	// 确定运行模式
	runMode := "stdio"
	if *sse {
		runMode = "sse"
	} else if *streamableHTTP {
		runMode = "streamable-http"
	}

	logger.Info("Starting server",
		zap.String("mode", runMode),
		zap.String("host", *host),
		zap.Int("port", *port),
		zap.String("path", *path))

	// 启动服务器
	var err error
	switch runMode {
	case "sse":
		err = server.ServeSSE(*host, *port, *path)
	case "streamable-http":
		err = server.ServeStreamableHTTP(*host, *port, *path)
	default:
		err = server.ServeStdio()
	}

	if err != nil {
		logger.Fatal("Server error", zap.Error(err))
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println(version.Info())
}

func handleToolsCommand(args []string) {
	if len(args) == 0 || args[0] != "list" {
		fmt.Fprintln(os.Stderr, "Usage: siriusec-mcp tools list [--format <format>]")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("tools list", flag.ExitOnError)
	format := fs.String("format", "simple", "Output format: simple/table/json")
	fs.Parse(args[1:])

	if err := cli.ListTools(*format); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleConfigCommand(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	configDir := fs.String("dir", ".", "Configuration directory")
	fs.Parse(args)

	if len(fs.Args()) == 0 || fs.Args()[0] != "validate" {
		fmt.Fprintln(os.Stderr, "Usage: siriusec-mcp config validate [flags]")
		os.Exit(1)
	}

	if err := cli.ValidateConfig(*configDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleSelfTestCommand(args []string) {
	fs := flag.NewFlagSet("selftest", flag.ExitOnError)
	filter := fs.String("filter", "", "Filter tools by name or category")
	verbose := fs.Bool("v", false, "Verbose output")
	openapi := fs.Bool("openapi", false, "Test OpenAPI connection")
	fs.Parse(args)

	if *openapi {
		if err := cli.TestOpenAPIConnection(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := cli.RunSelfTest(*filter, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleMCPTestCommand(args []string) {
	fs := flag.NewFlagSet("mcptest", flag.ExitOnError)
	baseURL := fs.String("url", "http://127.0.0.1:7140", "MCP server base URL")
	fs.Parse(args)

	results := cli.RunMCPTest(*baseURL)
	cli.PrintMCPTestResults(results)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Siriusec MCP Server %s

Usage:
  siriusec-mcp run [flags]                    Start the MCP server (default)
  siriusec-mcp version                         Show version information
  siriusec-mcp help                            Show this help message
  siriusec-mcp tools list [flags]              List all available tools
  siriusec-mcp config validate [flags]         Validate configuration
  siriusec-mcp selftest [flags]                Run self-test
  siriusec-mcp mcptest [flags]                 Test MCP server connection

Run Flags:
  --sse                  Run in SSE mode
  --streamable-http      Run in streamable-http mode
  --host <host>          Host to bind to (default: 127.0.0.1)
  --port <port>          Port to bind to (default: 7140)
  --path <path>          Endpoint path (default: /mcp/unified)

Tools List Flags:
  --format <format>      Output format: simple/table/json (default: simple)

Config Validate Flags:
  --dir <dir>            Configuration directory (default: .)

Selftest Flags:
  --filter <filter>      Filter tools by name or category
  -v                     Verbose output
  --openapi              Test OpenAPI connection

MCP Test Flags:
  --url <url>            MCP server base URL (default: http://127.0.0.1:7140)

Examples:
  siriusec-mcp                              # Run in stdio mode
  siriusec-mcp run --sse --port 7140        # Run in SSE mode
  siriusec-mcp run --streamable-http        # Run in HTTP mode
  siriusec-mcp version                      # Show version
  siriusec-mcp tools list                   # List all tools
  siriusec-mcp tools list --format json     # List tools as JSON
  siriusec-mcp config validate              # Validate configuration
  siriusec-mcp selftest                     # Run all self-tests
  siriusec-mcp selftest --filter am         # Test AM tools only
  siriusec-mcp mcptest                      # Test local MCP server
  siriusec-mcp mcptest --url http://host:7140  # Test remote server

Environment Variables:
  ACCESS_KEY_ID          Aliyun AccessKey ID
  ACCESS_KEY_SECRET      Aliyun AccessKey Secret
  REGION_ID              Aliyun Region (default: cn-hangzhou)
  LOG_LEVEL              Log level: DEBUG/INFO/WARN/ERROR
`, version.Short())
}
