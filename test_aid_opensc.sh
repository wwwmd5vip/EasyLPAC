#!/bin/bash
# 使用 opensc-tool 测试和读取 AID 的脚本

echo "=== OpenSC 工具 - AID 读取脚本 ==="
echo ""

# 检查是否有读卡器
echo "1. 检查读卡器..."
READERS=$(opensc-tool --list-readers 2>&1)
if [ $? -eq 0 ]; then
    echo "$READERS"
else
    echo "错误: 无法列出读卡器"
    exit 1
fi

echo ""
echo "2. 等待卡片插入（按 Ctrl+C 取消）..."
opensc-tool --wait 2>&1

echo ""
echo "3. 读取卡片信息..."
echo "--- ATR (Answer To Reset) ---"
ATR=$(opensc-tool --atr 2>&1 | grep -v "Using reader")
echo "$ATR"

echo ""
echo "--- 尝试自动识别卡片类型 ---"
opensc-tool --name 2>&1 | grep -v "Using reader" || echo "无法自动识别，尝试指定驱动..."

echo ""
echo "--- 可用的卡片驱动 ---"
opensc-tool --card-driver ? 2>&1 | head -20

echo ""
echo "--- 尝试使用 default 驱动 ---"
opensc-tool -c default --name 2>&1 | grep -v "Using reader" || echo "default 驱动失败"

echo ""
echo "--- 尝试使用 cardos 驱动（如果可用）---"
opensc-tool -c cardos --name 2>&1 | grep -v "Using reader" || echo "cardos 驱动不可用或失败"

echo ""
echo "--- 尝试使用 atrust-acos 驱动（如果可用）---"
opensc-tool -c atrust-acos --name 2>&1 | grep -v "Using reader" || echo "atrust-acos 驱动不可用或失败"

echo ""
echo "4. 列出卡片上的文件（使用 default 驱动）..."
opensc-tool -c default --list-files 2>&1 | grep -v "Using reader" || echo "无法列出文件"

echo ""
echo "5. 测试常见 eUICC AID（使用 default 驱动）..."
echo "--- 测试 A0000005591010FFFFFFFF8900000100 (默认 eUICC ISD-R) ---"
opensc-tool -c default --send-apdu "00A4040010A0000005591010FFFFFFFF8900000100" 2>&1 | grep -v "Using reader"

echo ""
echo "--- 测试 A000000559 (eUICC 基础) ---"
opensc-tool -c default --send-apdu "00A4040005A000000559" 2>&1 | grep -v "Using reader"

echo ""
echo "6. 测试常见 USIM AID（使用 default 驱动）..."
echo "--- 测试 A0000000871002FF49FF0589 (USIM) ---"
opensc-tool -c default --send-apdu "00A404000BA0000000871002FF49FF0589" 2>&1 | grep -v "Using reader"

echo ""
echo "7. 使用 lpac 工具读取（推荐，因为 lpac 已配置好 PCSC 驱动）..."
LPAC_CMD=""
if command -v lpac &> /dev/null; then
    LPAC_CMD="lpac"
elif [ -f "./lpac" ]; then
    LPAC_CMD="./lpac"
fi

if [ -n "$LPAC_CMD" ]; then
    echo "--- 使用 lpac chip info 读取卡片信息 ---"
    $LPAC_CMD chip info 2>&1 | head -20 || echo "lpac 命令失败"
    
    echo ""
    echo "--- 使用 lpac profile list 列出应用 ---"
    $LPAC_CMD profile list 2>&1 | head -20 || echo "lpac profile list 失败（可能不是 eUICC 卡）"
else
    echo "lpac 未找到，请确保 lpac 在当前目录或 PATH 中"
    echo "或者使用 EasyLPAC 应用程序来读取 AID"
fi

echo ""
echo "8. 提示："
echo "   - OpenSC 可能无法识别某些卡片类型（如 eUICC）"
echo "   - lpac 工具使用 PCSC 直接通信，更适合 eUICC 卡"
echo "   - 建议使用 EasyLPAC 应用程序的'测试 AID'功能"
echo "   - 或者直接使用 lpac 命令行工具"

echo ""
echo "=== 完成 ==="
echo ""
echo "提示:"
echo "  - 如果看到 9000，表示 AID 存在"
echo "  - 如果看到 6A82，表示 AID 不存在"
echo "  - 如果看到其他错误，可能是卡片类型不支持或需要认证"

