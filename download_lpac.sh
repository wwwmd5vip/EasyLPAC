#!/bin/bash

# 下载lpac可执行文件的脚本

cd "$(dirname "$0")"

echo "=========================================="
echo "下载 lpac 可执行文件"
echo "=========================================="
echo ""

# 检测系统架构
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

echo "检测到系统: $OS $ARCH"
echo ""

# 根据系统选择下载链接
if [ "$OS" = "darwin" ]; then
    if [ "$ARCH" = "arm64" ]; then
        echo "macOS ARM64 (Apple Silicon)"
        DOWNLOAD_URL="https://github.com/estkme-group/lpac/releases/latest/download/lpac-darwin-arm64"
        FILENAME="lpac"
    elif [ "$ARCH" = "x86_64" ]; then
        echo "macOS x86_64 (Intel)"
        DOWNLOAD_URL="https://github.com/estkme-group/lpac/releases/latest/download/lpac-darwin-amd64"
        FILENAME="lpac"
    else
        echo "不支持的架构: $ARCH"
        exit 1
    fi
elif [ "$OS" = "linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        echo "Linux x86_64"
        DOWNLOAD_URL="https://github.com/estkme-group/lpac/releases/latest/download/lpac-linux-amd64"
        FILENAME="lpac"
    else
        echo "不支持的架构: $ARCH"
        exit 1
    fi
else
    echo "不支持的操作系统: $OS"
    exit 1
fi

echo ""
echo "下载URL: $DOWNLOAD_URL"
echo "保存为: $FILENAME"
echo ""

# 下载文件
if curl -L -o "$FILENAME" "$DOWNLOAD_URL" 2>&1; then
    # 添加执行权限
    chmod +x "$FILENAME"
    
    echo ""
    echo "=========================================="
    echo "下载成功！"
    echo "=========================================="
    echo ""
    echo "文件位置: $(pwd)/$FILENAME"
    echo "文件大小: $(ls -lh "$FILENAME" | awk '{print $5}')"
    echo ""
    echo "验证文件..."
    if file "$FILENAME" | grep -q "executable"; then
        echo "✓ 文件验证通过，是可执行文件"
    else
        echo "⚠ 警告: 文件可能不是可执行文件"
    fi
    echo ""
    echo "现在可以运行 EasyLPAC 了！"
else
    echo ""
    echo "=========================================="
    echo "下载失败！"
    echo "=========================================="
    echo ""
    echo "请手动下载 lpac:"
    echo "1. 访问: https://github.com/estkme-group/lpac/releases/latest"
    echo "2. 下载适合你系统的版本"
    echo "3. 将文件重命名为 'lpac' 并放在当前目录"
    echo "4. 运行: chmod +x lpac"
    exit 1
fi

