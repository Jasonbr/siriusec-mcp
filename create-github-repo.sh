#!/bin/bash
# 创建 GitHub 仓库脚本

REPO_NAME="siriusec-mcp"
DESCRIPTION="Siriusec MCP Server - 统一的 MCP 协议网关，集成阿里云 SysOM 监控服务"

# 检查是否有 GitHub CLI
if command -v gh &> /dev/null; then
    echo "使用 GitHub CLI 创建仓库..."
    gh repo create "$REPO_NAME" --public --description "$DESCRIPTION" --source=. --remote=origin --push
else
    echo "GitHub CLI 未安装，请手动创建仓库:"
    echo ""
    echo "1. 打开浏览器访问: https://github.com/new"
    echo ""
    echo "2. 填写以下信息:"
    echo "   Repository name: $REPO_NAME"
    echo "   Description: $DESCRIPTION"
    echo "   Visibility: Public"
    echo "   不要勾选 'Initialize this repository with a README'"
    echo ""
    echo "3. 点击 'Create repository'"
    echo ""
    echo "4. 创建完成后，运行以下命令推送代码:"
    echo "   git push -u origin main"
    echo ""
fi
