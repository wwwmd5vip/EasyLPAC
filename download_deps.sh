#!/bin/bash

# Go依赖下载脚本（带进度显示）
cd "$(dirname "$0")"

# 设置代理（使用多个镜像源）
export GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,https://goproxy.io,direct
export GOSUMDB=off

echo "=========================================="
echo "Go依赖下载工具（带进度显示）"
echo "=========================================="
echo ""
echo "当前Go代理: $GOPROXY"
echo "Go版本: $(go version)"
echo ""

# 方法1: 使用go get逐个下载（显示详细进度）
echo "方法1: 逐个下载主要依赖（显示详细进度）..."
echo "----------------------------------------"

# 从go.mod提取主要依赖
main_deps=(
    "fyne.io/fyne/v2@v2.6.0"
    "github.com/Xuanwo/go-locale@v1.1.3"
    "github.com/fullpipe/icu-mf@v1.0.1"
    "github.com/makiuchi-d/gozxing@v0.1.1"
    "github.com/mattn/go-runewidth@v0.0.16"
    "github.com/sqweek/dialog@latest"
    "golang.design/x/clipboard@v0.7.0"
    "golang.org/x/net@v0.39.0"
    "golang.org/x/text@v0.24.0"
)

total=${#main_deps[@]}
count=0

for dep in "${main_deps[@]}"; do
    count=$((count + 1))
    echo ""
    echo "[$count/$total] 正在下载: $dep"
    echo "----------------------------------------"
    
    # 使用go get -v显示详细输出
    if go get -v "$dep" 2>&1 | tee /tmp/go_download.log; then
        echo "✓ 下载成功: $dep"
    else
        echo "✗ 下载失败: $dep"
        cat /tmp/go_download.log | tail -5
    fi
done

echo ""
echo "----------------------------------------"
echo "方法2: 下载所有依赖（包括间接依赖）..."
echo "----------------------------------------"

# 使用go mod download，但通过管道显示进度
go mod download 2>&1 | while IFS= read -r line; do
    if [[ "$line" =~ downloading|get|go: ]]; then
        echo "$line"
    fi
done

echo ""
echo "=========================================="
echo "验证依赖..."
echo "=========================================="

# 验证依赖是否完整
if go mod verify 2>&1; then
    echo ""
    echo "✓ 所有依赖验证通过！"
else
    echo ""
    echo "⚠ 部分依赖验证失败，但可能仍可使用"
fi

echo ""
echo "=========================================="
echo "下载完成！"
echo "=========================================="
echo ""
echo "可以使用以下命令编译："
echo "  go build -o EasyLPAC ."
echo ""
# cd /Users/takj/Downloads/Github/EasyLPAC

# # 下载依赖
# go mod download

# # 编译
# go build -o EasyLPAC .

# # 运行
# ./EasyLPAC

# go run .