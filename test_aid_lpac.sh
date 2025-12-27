#!/bin/bash
# 使用 lpac 工具测试和读取 AID 的脚本（推荐方法）

echo "=== LPAC 工具 - AID 读取脚本 ==="
echo ""

# 查找 lpac 可执行文件
LPAC_CMD=""
if command -v lpac &> /dev/null; then
    LPAC_CMD="lpac"
elif [ -f "./lpac" ]; then
    LPAC_CMD="./lpac"
elif [ -f "$(dirname "$0")/lpac" ]; then
    LPAC_CMD="$(dirname "$0")/lpac"
else
    echo "错误: lpac 未找到"
    echo "请确保 lpac 在当前目录或 PATH 中"
    echo "或者使用 EasyLPAC 应用程序"
    exit 1
fi

echo "使用 lpac: $LPAC_CMD"
echo "lpac 版本:"
$LPAC_CMD version 2>&1 | head -5

echo ""
echo "1. 读取卡片基本信息（使用默认 AID）..."
echo "--- chip info ---"
$LPAC_CMD chip info 2>&1 | head -30

echo ""
echo "2. 列出 Profile（如果是 eUICC 卡，使用默认 AID）..."
echo "--- profile list ---"
$LPAC_CMD profile list 2>&1 | head -20

echo ""
echo "3. 测试不同的 AID..."
echo "--- 测试 eUICC 默认 AID: A0000005591010FFFFFFFF8900000100 ---"
LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900000100" $LPAC_CMD chip info 2>&1 | head -10

echo ""
echo "--- 测试 eUICC 5BER AID: A0000005591010FFFFFFFF8900050500 ---"
LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900050500" $LPAC_CMD chip info 2>&1 | head -10

echo ""
echo "--- 测试 eSIM.me AID: A0000005591010000000008900000300 ---"
LPAC_CUSTOM_ISD_R_AID="A0000005591010000000008900000300" $LPAC_CMD chip info 2>&1 | head -10

echo ""
echo "4. 列出可用的 APDU 驱动..."
echo "--- driver apdu list ---"
$LPAC_CMD driver apdu list 2>&1 | head -20

echo ""
echo "=== 完成 ==="
echo ""
echo "提示:"
echo "  - 如果 chip info 返回 code:0，说明卡片连接正常且 AID 正确"
echo "  - 如果返回 code:-1 和 'euicc_init'，说明 AID 不正确或卡片不是 eUICC"
echo "  - 如果看到 'Sharing violation' 错误，说明读卡器被其他程序占用："
echo "    1. 关闭 EasyLPAC 应用程序（如果正在运行）"
echo "    2. 关闭其他使用读卡器的程序"
echo "    3. 等待几秒后重试"
echo "  - 如果失败，可以："
echo "    1. 使用 EasyLPAC 应用程序的'测试 AID'功能自动查找正确的 AID"
echo "    2. 手动设置 AID 环境变量："
echo "       export LPAC_CUSTOM_ISD_R_AID='你的AID'"
echo "       $LPAC_CMD chip info"
echo "    3. 检查 aid.txt 文件中的 AID 列表，尝试不同的 AID"

