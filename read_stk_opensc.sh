#!/bin/bash
# 使用 OpenSC 工具读取 SIM 卡上的 STK 应用程序

echo "=== 使用 OpenSC 读取 STK 应用程序 ==="
echo ""

# 检查 OpenSC 是否安装
if ! command -v opensc-tool &> /dev/null; then
    echo "错误: opensc-tool 未找到"
    echo "请安装 OpenSC:"
    echo "  macOS: brew install opensc"
    echo "  Linux: sudo apt-get install opensc"
    exit 1
fi

echo "1. 检查读卡器..."
READERS=$(opensc-tool --list-readers 2>&1)
if [ $? -eq 0 ]; then
    echo "$READERS"
else
    echo "错误: 无法列出读卡器"
    exit 1
fi

echo ""
echo "2. 读取卡片信息..."
echo "--- ATR ---"
opensc-tool --atr 2>&1

echo ""
echo "--- 卡片名称 ---"
opensc-tool --name 2>&1

echo ""
echo "3. 列出卡片上的文件..."
echo "--- 列出所有文件 ---"
opensc-tool --list-files 2>&1 | head -50

echo ""
echo "4. 尝试读取应用目录文件 (EF_DIR)..."
# EF_DIR 通常在 2F00
echo "--- 读取 2F00 (EF_DIR) ---"
opensc-tool --send-apdu "00A40000022F00" 2>&1
opensc-tool --send-apdu "00B0000000" 2>&1

echo ""
echo "5. 尝试读取应用信息..."
echo "--- GET DATA: Application Template ---"
opensc-tool --send-apdu "00CA004500" 2>&1

echo ""
echo "--- GET DATA: Application Label ---"
opensc-tool --send-apdu "00CA004300" 2>&1

echo ""
echo "6. 尝试选择并读取 STK 应用..."
echo "--- 选择 ISIM/STK AID: A0000000871004FF86FF4989 ---"
opensc-tool --send-apdu "00A404000CA0000000871004FF86FF4989" 2>&1

echo ""
echo "--- 选择 USIM/STK AID: A0000000871002FF86FF4989 ---"
opensc-tool --send-apdu "00A404000CA0000000871002FF86FF4989" 2>&1

echo ""
echo "7. 使用 pkcs15-tool 列出应用（如果可用）..."
if command -v pkcs15-tool &> /dev/null; then
    echo "--- 列出所有应用 ---"
    pkcs15-tool --list-applications 2>&1 | head -30
    
    echo ""
    echo "--- 列出所有文件 ---"
    pkcs15-tool --list-files 2>&1 | head -50
else
    echo "pkcs15-tool 未安装，跳过"
fi

echo ""
echo "=== 完成 ==="
echo ""
echo "提示:"
echo "  - 如果看到 9000，表示命令成功"
echo "  - 如果看到 6A82，表示文件或应用不存在"
echo "  - 使用 Python 脚本 read_stk_applet.py 可以获得更详细的信息"

