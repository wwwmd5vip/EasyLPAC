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
echo "检查读卡器状态..."
if $LPAC_CMD driver apdu list 2>&1 | grep -q "Generic\|Reader"; then
    echo "✓ 读卡器已检测到"
else
    echo "✗ 未检测到读卡器，请检查连接"
    exit 1
fi

echo ""
echo "检查是否有 EasyLPAC 进程正在运行..."
EASYLPAC_PID=$(ps aux | grep -iE "easylpac|EasyLPAC" | grep -v grep | awk '{print $2}' | head -1)
if [ -n "$EASYLPAC_PID" ]; then
    echo "⚠️  发现 EasyLPAC 进程正在运行 (PID: $EASYLPAC_PID)"
    echo "   这会导致 'Sharing violation' 错误"
    echo ""
    echo "   解决方案："
    echo "   1. 关闭 EasyLPAC 应用程序（推荐：直接使用应用程序的'测试 AID'功能）"
    echo "   2. 或者等待 EasyLPAC 释放读卡器后重试此脚本"
    echo ""
    echo "   提示：如果 EasyLPAC 正在运行，建议直接使用应用程序内的'测试 AID'功能，"
    echo "         这样更可靠且不需要关闭应用程序。"
    echo ""
else
    echo "✓ 未发现 EasyLPAC 进程"
fi

echo ""
echo "注意: 如果看到 'Sharing violation' 错误，请："
echo "  1. 关闭 EasyLPAC 应用程序（如果正在运行）"
echo "  2. 等待 3-5 秒让 PCSC 服务释放资源"
echo "  3. 如果问题持续，尝试重启 PCSC 服务："
echo "     sudo launchctl stop org.opensc.pcscd"
echo "     sudo launchctl start org.opensc.pcscd"

echo ""
echo "1. 读取卡片基本信息（使用默认 AID）..."
echo "--- chip info ---"
CHIP_INFO_OUTPUT=$($LPAC_CMD chip info 2>&1)
echo "$CHIP_INFO_OUTPUT" | head -30

# 检查是否有 Sharing violation 错误
if echo "$CHIP_INFO_OUTPUT" | grep -q "Sharing violation"; then
    echo ""
    echo "⚠️  检测到 'Sharing violation' 错误！"
    echo "   读卡器被其他程序占用，请："
    echo "   1. 关闭 EasyLPAC 应用程序（如果正在运行）"
    echo "   2. 关闭其他使用读卡器的程序"
    echo "   3. 等待 3-5 秒后重新运行此脚本"
    echo ""
    echo "   或者直接使用 EasyLPAC 应用程序的'测试 AID'功能"
    echo ""
fi

echo ""
echo "2. 列出 Profile（如果是 eUICC 卡，使用默认 AID）..."
echo "--- profile list ---"
$LPAC_CMD profile list 2>&1 | head -20

# 如果检测到 Sharing violation，跳过 AID 测试
if ! echo "$CHIP_INFO_OUTPUT" | grep -q "Sharing violation"; then
    echo ""
    echo "3. 测试不同的 AID..."
    echo ""
    echo "--- 测试 eUICC AID（如果是 eUICC 卡）---"
    echo "测试 eUICC 默认 AID: A0000005591010FFFFFFFF8900000100"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900000100" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 eUICC AID!"
    fi

    echo ""
    echo "测试 eUICC 5BER AID: A0000005591010FFFFFFFF8900050500"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900050500" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 eUICC AID!"
    fi

    echo ""
    echo "测试 eSIM.me AID: A0000005591010000000008900000300"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000005591010000000008900000300" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 eUICC AID!"
    fi

    echo ""
    echo "--- 测试传统 SIM 卡 AID（如果是普通手机卡）---"
    echo "测试 USIM AID: A0000000871002FF49FF0589"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871002FF49FF0589" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 USIM AID!"
    fi

    echo ""
    echo "测试 USIM AID (短): A0000000871002"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871002" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 USIM AID!"
    fi

    echo ""
    echo "测试 USIM AID (基础): A000000087"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A000000087" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 USIM AID!"
    fi

    echo ""
    echo "--- 提示：如果上述 AID 都失败 ---"
    echo "传统 SIM 卡可能需要特定的完整 AID，不是基础前缀"
    echo "建议使用 EasyLPAC 应用程序的'测试 AID'功能，"
    echo "它会自动测试 aid.txt 中的所有 AID（包括传统 SIM 卡的各种 AID）"

    echo ""
    echo "测试 SIM AID: A0000000030000"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000030000" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 SIM AID!"
    fi

    echo ""
    echo "测试 ISIM AID: A0000000871004FF86FF4989"
    TEST_OUTPUT=$(LPAC_CUSTOM_ISD_R_AID="A0000000871004FF86FF4989" $LPAC_CMD chip info 2>&1)
    echo "$TEST_OUTPUT" | head -5
    if echo "$TEST_OUTPUT" | grep -q '"code":0'; then
        echo "✓ 找到有效的 ISIM AID!"
    fi
else
    echo ""
    echo "3. 跳过 AID 测试（读卡器被占用）"
    echo "   请先解决 'Sharing violation' 错误"
fi

echo ""
echo "4. 列出可用的 APDU 驱动..."
echo "--- driver apdu list ---"
$LPAC_CMD driver apdu list 2>&1 | head -20

echo ""
echo "=== 完成 ==="
echo ""
echo "提示:"
echo "  - 如果 chip info 返回 code:0，说明卡片连接正常且 AID 正确"
echo "  - 如果返回 code:-1 和 'euicc_init'，说明 AID 不正确"
echo "  - 卡片类型："
echo "    * eUICC 卡：通常以 A000000559 开头的 AID"
echo "    * 传统 SIM 卡（移动/联通/电信）：通常以 A000000087 开头的 AID（USIM）"
echo "    * 传统 SIM 卡：也可能使用 A000000003 开头的 AID（SIM）"
echo "  - 如果看到 'Sharing violation' 错误，说明读卡器被其他程序占用："
echo "    1. 关闭 EasyLPAC 应用程序（如果正在运行）"
echo "    2. 关闭其他使用读卡器的程序"
echo "    3. 等待几秒后重试"
echo "  - 如果所有测试都失败，可以："
echo "    1. 使用 EasyLPAC 应用程序的'测试 AID'功能自动查找正确的 AID"
echo "       （应用程序会测试 aid.txt 中的所有 AID，包括传统 SIM 卡 AID）"
echo "    2. 手动设置 AID 环境变量："
echo "       export LPAC_CUSTOM_ISD_R_AID='你的AID'"
echo "       $LPAC_CMD chip info"
echo "    3. 检查 aid.txt 文件中的 AID 列表，尝试不同的 AID"
echo "    4. 如果是传统 SIM 卡，尝试以 A000000087 或 A000000003 开头的 AID"

