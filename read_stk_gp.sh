#!/bin/bash
# 使用 GlobalPlatform Pro 工具读取 STK 应用程序的脚本

echo "=== 使用 GlobalPlatform Pro 读取 STK 应用程序 ==="
echo ""

# 检查 Java 是否安装
if ! command -v java &> /dev/null; then
    echo "错误: Java 未找到"
    echo "请安装 Java:"
    echo "  macOS: brew install openjdk"
    echo "  Linux: sudo apt-get install default-jdk"
    exit 1
fi

# 查找 gp.jar
GP_JAR=""
if [ -f "./gp.jar" ]; then
    GP_JAR="./gp.jar"
elif [ -f "$(dirname "$0")/gp.jar" ]; then
    GP_JAR="$(dirname "$0")/gp.jar"
else
    echo "错误: gp.jar 未找到"
    echo "请确保 gp.jar 在当前目录或下载 GlobalPlatform Pro:"
    echo "  https://github.com/martinpaljak/GlobalPlatformPro/releases"
    exit 1
fi

echo "使用 GlobalPlatform Pro: $GP_JAR"
echo ""

# 检查是否有 EasyLPAC 进程
echo "1. 检查是否有 EasyLPAC 进程..."
EASYLPAC_PID=$(ps aux | grep -iE "easylpac|EasyLPAC" | grep -v grep | awk '{print $2}' | head -1)
if [ -n "$EASYLPAC_PID" ]; then
    echo "⚠️  发现 EasyLPAC 进程正在运行 (PID: $EASYLPAC_PID)"
    echo "   建议关闭 EasyLPAC 后重试，或直接使用 EasyLPAC 应用程序"
    echo ""
fi

# 列出所有应用
echo "2. 列出卡片上的所有应用..."
echo "--- 使用 gp -l 列出应用 ---"

# 尝试不同的方法连接
SUCCESS=0
APPLICATIONS=""

# 方法1: 直接列出
echo "尝试方法1: 直接连接..."
APPLICATIONS=$(java -jar "$GP_JAR" -l 2>&1)
if [ $? -eq 0 ] && ! echo "$APPLICATIONS" | grep -q "Error\|SCardConnect"; then
    echo "$APPLICATIONS"
    SUCCESS=1
else
    echo "方法1失败，尝试方法2..."
    
    # 方法2: 使用独占模式
    echo ""
    echo "尝试方法2: 使用独占模式 (--pcsc-exclusive)..."
    APPLICATIONS=$(java -jar "$GP_JAR" -l --pcsc-exclusive 2>&1)
    if [ $? -eq 0 ] && ! echo "$APPLICATIONS" | grep -q "Error\|SCardConnect"; then
        echo "$APPLICATIONS"
        SUCCESS=1
    else
        # 方法3: 指定读卡器
        echo ""
        echo "尝试方法3: 指定读卡器..."
        READER_NAME=""
        
        # 尝试使用 pcsc_scan 检测读卡器（如果可用）
        if command -v pcsc_scan &> /dev/null; then
            READER_NAME=$(pcsc_scan 2>&1 | grep -i "reader" | head -1 | awk -F: '{print $2}' | xargs)
        fi
        
        if [ -n "$READER_NAME" ]; then
            echo "检测到读卡器: $READER_NAME"
            APPLICATIONS=$(java -jar "$GP_JAR" -l --reader "$READER_NAME" 2>&1)
            if [ $? -eq 0 ] && ! echo "$APPLICATIONS" | grep -q "Error\|SCardConnect"; then
                echo "$APPLICATIONS"
                SUCCESS=1
            fi
        else
            # 尝试常见的读卡器名称
            for READER in "Generic EMV Smartcard Reader" "ACS ACR122U" "OMNIKEY"; do
                echo "尝试读卡器: $READER"
                APPLICATIONS=$(java -jar "$GP_JAR" -l --reader "$READER" 2>&1)
                if [ $? -eq 0 ] && ! echo "$APPLICATIONS" | grep -q "Error\|SCardConnect"; then
                    echo "$APPLICATIONS"
                    SUCCESS=1
                    break
                fi
            done
        fi
    fi
fi

if [ $SUCCESS -eq 0 ]; then
    echo ""
    echo "⚠️  无法连接到卡片，错误信息："
    echo "$APPLICATIONS" | grep -E "Error|SCardConnect" | head -3
    echo ""
    echo "可能的原因和解决方案："
    echo ""
    echo "1. 读卡器被其他程序占用："
    echo "   - 关闭 EasyLPAC 应用程序（如果正在运行）"
    echo "   - 等待 3-5 秒让 PCSC 服务释放资源"
    echo "   - 检查是否有其他程序在使用读卡器"
    echo ""
    echo "2. 卡片未正确插入或接触不良："
    echo "   - 检查卡片是否正确插入"
    echo "   - 尝试重新插入卡片"
    echo "   - 清洁卡片和读卡器触点"
    echo ""
    echo "3. PCSC 服务问题（macOS）："
    echo "   sudo launchctl stop org.opensc.pcscd"
    echo "   sudo launchctl start org.opensc.pcscd"
    echo ""
    echo "4. 卡片需要先选择正确的 AID："
    echo "   - 某些卡片可能需要先使用 lpac 工具选择 AID"
    echo "   - 尝试使用 EasyLPAC 应用程序的'测试 AID'功能"
    echo ""
    echo "5. 使用调试模式查看详细信息："
    echo "   java -jar $GP_JAR -d -l"
    echo ""
    echo "替代方案："
    echo "  - 使用 read_stk_applet.py（Python 脚本）"
    echo "  - 使用 read_stk_opensc.sh（OpenSC 工具）"
    echo "  - 使用 EasyLPAC 应用程序的'测试 AID'功能"
    echo ""
    exit 1
fi

echo ""
echo "3. 尝试读取常见 STK AID 的应用信息..."
echo ""

# 常见的 STK AID（用于测试）
STK_AIDS=(
    "A0000000871004FF86FF4989"  # ISIM/STK
    "A0000000871002FF86FF4989"  # USIM/STK
    "A0000000871002FF49FF0589"  # USIM
    "A000000087"                 # 3GPP 基础
    "A0000000030000"             # SIM
    "A0000005591010FFFFFFFF8900000100"  # eUICC
    "A0000005591010FFFFFFFF8900000177"  # eUICC (用户修改的)
)

for AID in "${STK_AIDS[@]}"; do
    echo "--- 应用信息: $AID ---"
    # 使用 --applet 或 -c 参数指定 AID，然后使用 --info
    java -jar "$GP_JAR" --applet "$AID" --info 2>&1 | head -20
    if [ $? -ne 0 ]; then
        # 如果失败，尝试使用 -c 连接
        java -jar "$GP_JAR" -c "$AID" --info 2>&1 | head -20
    fi
    echo ""
done

echo ""
echo "4. 尝试获取应用数据（如果支持）..."
echo ""

# 尝试获取应用数据
for AID in "${STK_AIDS[@]}"; do
    echo "--- 获取数据: $AID ---"
    java -jar "$GP_JAR" --applet "$AID" --get-data 2>&1 | head -10
    if [ $? -ne 0 ]; then
        java -jar "$GP_JAR" -c "$AID" --get-data 2>&1 | head -10
    fi
    echo ""
done

echo ""
echo "=== 完成 ==="
echo ""
echo "提示："
echo "  - 如果看到 'No card found'，请检查读卡器连接"
echo "  - 如果看到 'Authentication failed'，需要 ADM 密钥"
echo "  - 某些操作可能需要管理员权限（ADM 密钥）"
echo "  - 可以使用 --reader 参数指定读卡器："
echo "    java -jar $GP_JAR -l --reader \"读卡器名称\""
echo ""
echo "GlobalPlatform Pro 常用命令："
echo "  java -jar $GP_JAR -l                              # 列出所有应用"
echo "  java -jar $GP_JAR --applet <AID> --info           # 获取应用信息"
echo "  java -jar $GP_JAR -c <AID> --info                 # 连接应用并获取信息"
echo "  java -jar $GP_JAR --applet <AID> --get-data       # 获取应用数据"
echo "  java -jar $GP_JAR --delete <AID>                  # 删除应用（需要 ADM 密钥）"
echo "  java -jar $GP_JAR --install <CAP文件>             # 安装应用（需要 ADM 密钥）"
echo "  java -jar $GP_JAR --reader \"读卡器名称\" -l        # 指定读卡器列出应用"
echo "  java -jar $GP_JAR -d -l                           # 调试模式列出应用"
echo ""
echo "注意："
echo "  - 某些操作需要 ADM 密钥（管理员密钥）"
echo "  - 如果读卡器被占用，请关闭 EasyLPAC 后重试"
echo "  - 可以使用 --reader 参数指定读卡器"
echo "  - 使用 -d 参数可以查看详细的 APDU 通信"
echo ""
