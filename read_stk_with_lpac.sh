#!/bin/bash
# 使用 lpac 工具读取 STK 应用程序的脚本
# 这个方法先使用 lpac 建立连接，然后读取应用数据

echo "=== 使用 lpac 工具读取 STK 应用程序 ==="
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
echo "1. 检查读卡器状态..."
if ! $LPAC_CMD driver apdu list 2>&1 | grep -q "Generic\|Reader"; then
    echo "✗ 未检测到读卡器，请检查连接"
    exit 1
fi
echo "✓ 读卡器已检测到"
echo ""

# 检查是否有 EasyLPAC 进程
echo "2. 检查是否有 EasyLPAC 进程..."
EASYLPAC_PID=$(ps aux | grep -iE "easylpac|EasyLPAC" | grep -v grep | awk '{print $2}' | head -1)
if [ -n "$EASYLPAC_PID" ]; then
    echo "⚠️  发现 EasyLPAC 进程正在运行 (PID: $EASYLPAC_PID)"
    echo "   建议关闭 EasyLPAC 后重试，或直接使用 EasyLPAC 应用程序"
    echo ""
fi

# 读取卡片基本信息
echo "3. 读取卡片基本信息..."
echo "--- 使用默认 AID ---"
CHIP_INFO=$($LPAC_CMD chip info 2>&1)
echo "$CHIP_INFO" | head -20

# 检查是否成功
if echo "$CHIP_INFO" | grep -q '"code":0'; then
    echo "✓ 卡片连接成功，可以读取信息"
    echo ""
    echo "完整的卡片信息："
    echo "$CHIP_INFO" | head -50
else
    echo "✗ 无法读取卡片信息，可能需要正确的 AID"
    echo ""
    echo "4. 尝试使用 EasyLPAC 应用程序的'测试 AID'功能找到正确的 AID"
    echo "   或者手动测试不同的 AID："
    echo ""
    echo "   测试 eUICC AID:"
    echo "   LPAC_CUSTOM_ISD_R_AID='A0000005591010FFFFFFFF8900000100' $LPAC_CMD chip info"
    echo ""
    echo "   测试 STK/ISIM AID:"
    echo "   LPAC_CUSTOM_ISD_R_AID='A0000000871004FF86FF4989' $LPAC_CMD chip info"
    echo ""
    echo "   测试 USIM AID:"
    echo "   LPAC_CUSTOM_ISD_R_AID='A0000000871002FF49FF0589' $LPAC_CMD chip info"
fi

echo ""
echo "=== 完成 ==="
echo ""
echo "提示："
echo "  - lpac 工具主要用于 eUICC 卡管理，读取传统 SIM 卡的 STK 程序功能有限"
echo "  - 要读取完整的 STK 程序代码，建议使用："
echo "    1. Python 脚本: python3 read_stk_applet.py（需要先解决连接问题）"
echo "    2. OpenSC 工具: ./read_stk_opensc.sh"
echo "    3. GlobalPlatform Pro: gp -l（列出应用）"
echo "  - 如果卡片是 eUICC，可以使用 lpac profile list 查看 Profile"

