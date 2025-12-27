#!/bin/bash
# 读取 STK (SIM Tool Kit) 程序的脚本

echo "=== STK (SIM Tool Kit) 程序读取脚本 ==="
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
    exit 1
fi

echo "使用 lpac: $LPAC_CMD"
echo ""

# 检查读卡器
echo "检查读卡器状态..."
if ! $LPAC_CMD driver apdu list 2>&1 | grep -q "Generic\|Reader"; then
    echo "✗ 未检测到读卡器，请检查连接"
    exit 1
fi
echo "✓ 读卡器已检测到"
echo ""

# STK 相关的常见 AID
echo "1. 测试 STK/ISIM AID..."
echo "--- 测试 A0000000871004FF86FF4989 (ISIM/STK) ---"
TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871004FF86FF4989" $LPAC_CMD chip info 2>&1)
echo "$TEST_OUTPUT" | head -10
if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
    echo "✓ 找到有效的 STK/ISIM AID: A0000000871004FF86FF4989"
    echo ""
    echo "使用此 AID 读取卡片信息："
    LPAC_CUSTOM_ISD_R_AID="A0000000871004FF86FF4989" $LPAC_CMD chip info 2>&1 | head -30
    exit 0
fi

echo ""
echo "--- 测试 A0000000871002FF86FF4989 (USIM/STK) ---"
TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871002FF86FF4989" $LPAC_CMD chip info 2>&1)
echo "$TEST_OUTPUT" | head -10
if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
    echo "✓ 找到有效的 STK/USIM AID: A0000000871002FF86FF4989"
    echo ""
    echo "使用此 AID 读取卡片信息："
    LPAC_CUSTOM_ISD_R_AID="A0000000871002FF86FF4989" $LPAC_CMD chip info 2>&1 | head -30
    exit 0
fi

echo ""
echo "2. 测试其他可能的 STK AID..."
echo "--- 测试 A0000000871004 (ISIM 基础) ---"
TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871004" $LPAC_CMD chip info 2>&1)
echo "$TEST_OUTPUT" | head -5
if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
    echo "✓ 找到有效的 ISIM AID: A0000000871004"
    exit 0
fi

echo ""
echo "--- 测试 A0000000871002 (USIM 基础) ---"
TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871002" $LPAC_CMD chip info 2>&1)
echo "$TEST_OUTPUT" | head -5
if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
    echo "✓ 找到有效的 USIM AID: A0000000871002"
    exit 0
fi

echo ""
echo "3. 从 aid.txt 中查找所有可能的 STK/ISIM AID..."
if [ -f "aid.txt" ]; then
    echo "找到以下可能的 STK/ISIM AID："
    grep -E "^A000000087.*1004|^A000000087.*FF86" aid.txt | head -10 || echo "未找到匹配的 AID"
    echo ""
    echo "提示：可以尝试使用 EasyLPAC 应用程序的'测试 AID'功能，"
    echo "      它会自动测试 aid.txt 中的所有 AID，包括 STK 相关的 AID。"
else
    echo "未找到 aid.txt 文件"
fi

echo ""
echo "=== 完成 ==="
echo ""
echo "提示："
echo "  - STK (SIM Tool Kit) 是 SIM 卡上的应用程序，用于提供增值服务"
echo "  - STK 应用程序通常使用以 A000000087 开头的 AID"
echo "  - 如果上述测试都失败，建议："
echo "    1. 使用 EasyLPAC 应用程序的'测试 AID'功能自动查找"
echo "    2. 检查 aid.txt 文件中的其他 AID"
echo "    3. 确认卡片是否支持 STK 功能"

