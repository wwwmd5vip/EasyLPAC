# AID 数据获取指南

## 当前状态

当前 `aid.txt` 文件包含 **4909** 个AID条目，涵盖了：
- eUICC相关AID（以 `A000000559` 开头）
- 传统SIM卡AID（如 `A000000087` 开头的USIM）
- 支付卡AID（Visa、MasterCard等）
- 其他智能卡应用AID

## 获取更多AID的方法

### 1. 使用 fetch_aids.go 工具

已创建 `fetch_aids.go` 工具，可以：
- 从网络URL获取AID列表
- 合并多个AID文件
- 自动去重并保留更详细的描述
- 自动备份现有文件

**使用方法：**
```bash
# 编译工具
go build -o fetch_aids fetch_aids.go

# 运行更新（需要先在代码中配置AID数据源URL）
./fetch_aids update
```

**配置AID数据源：**
编辑 `fetch_aids.go`，在 `aidSources` 数组中添加AID数据源URL：
```go
aidSources := []string{
    "https://example.com/aid-list.txt",
    // 添加更多URL...
}
```

### 2. 手动添加AID

直接编辑 `aid.txt` 文件，格式为：
```
AID: 描述
```

例如：
```
A0000005591010FFFFFFFF8900000100: eUICC ISD-R Default
A0000000871002FF49FF0589: Telenor USIM
```

### 3. AID数据源推荐

#### 官方标准文档
- **ISO/IEC 7816-5**: AID注册标准
- **ETSI TS 102 221**: SIM/USIM应用标识符
- **3GPP TS 31.102**: USIM应用规范
- **GSMA SGP.22**: eUICC规范（包含ISD-R AID）

#### 在线资源
- **EMVCo**: 支付卡AID注册表
- **GlobalPlatform**: 智能卡平台AID
- **GitHub**: 搜索 "smart card AID list" 或 "ISO 7816 AID"

#### 工具读取
使用专业工具从实际卡片中读取AID：

**macOS:**
- **OpenSC**: `brew install opensc`，然后使用 `opensc-tool --list-aids`
- **pcsc-lite**: macOS 通常已预装，但需要安装工具：
  ```bash
  # 安装 OpenSC（包含 opensc-tool）
  brew install opensc
  
  # 列出卡片上的所有AID
  opensc-tool --list-aids
  
  # 或者使用 Python pcsc-tools（如果可用）
  pip install pyscard
  ```

**Linux:**
- **pcsc-tools**: `sudo apt-get install pcsc-tools`，然后使用 `pcsc_scan`
- **opensc-tools**: `sudo apt-get install opensc`，然后使用 `opensc-tool --list-aids`

**Windows:**
- **OpenSC**: 下载安装包，使用 `opensc-tool.exe --list-aids`
- **pcsc-tools**: 需要从源码编译或使用预编译版本

### 4. 从实际卡片读取AID

**macOS 方法：**

```bash
# 安装 OpenSC（包含 opensc-tool）
brew install opensc

# 方法1: 列出卡片上的文件（可能包含AID信息）
opensc-tool --list-files

# 方法2: 使用 Python pyscard 读取AID（推荐）
pip install pyscard
python3 << EOF
from smartcard.System import readers
from smartcard.util import toHexString

# 获取读卡器
r = readers()
if r:
    connection = r[0].createConnection()
    connection.connect()
    
    # 使用 SELECT 命令枚举应用（需要根据卡片类型调整）
    # 这是一个示例，实际命令可能因卡片而异
    SELECT = [0x00, 0xA4, 0x04, 0x00, 0x00]  # SELECT by name
    data, sw1, sw2 = connection.transmit(SELECT)
    print(f"Response: {toHexString(data)}, Status: {hex(sw1)}{hex(sw2)}")
else:
    print("未找到读卡器")
EOF

# 方法3: 查看卡片信息
opensc-tool --info
opensc-tool --atr
opensc-tool --name
opensc-tool --serial

# 方法4: 使用提供的测试脚本（推荐）
./test_aid_opensc.sh

# 方法5: 手动测试特定AID
# 格式: opensc-tool --send-apdu "CLA INS P1 P2 Lc [AID字节]"
# 例如测试 eUICC AID:
opensc-tool --send-apdu "00A4040010A0000005591010FFFFFFFF8900000100"
# 返回 9000 表示成功，6A82 表示不存在
```

**注意：** 
- `opensc-tool` 没有直接的 `--list-aids` 选项，需要使用 APDU 命令或 Python 脚本来枚举应用
- 已提供 `test_aid_opensc.sh` 脚本用于快速测试常见 AID
- APDU 命令格式：`00A40400[长度][AID字节]`，其中：
  - `00A40400` 是 SELECT 命令
  - 长度是 AID 的字节数（十六进制）
  - 后面是 AID 的十六进制字节

**Linux 方法：**

```bash
# 安装pcsc-tools
sudo apt-get install pcsc-tools

# 扫描卡片上的所有AID
pcsc_scan

# 或者使用 opensc-tools
sudo apt-get install opensc
opensc-tool --list-files
```

**使用提供的 Python 脚本：**

```bash
# 安装依赖
pip install pyscard

# 运行脚本读取AID
python3 list_aids.py
```

**使用 lpac 工具（需要手动设置 AID）：**

```bash
# 使用提供的脚本（会自动测试多个常见 AID）
./test_aid_lpac.sh

# 或手动使用 lpac 命令（需要设置正确的 AID）
# 方法1: 使用环境变量设置 AID
export LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900000100"
lpac chip info          # 读取卡片信息
lpac profile list       # 列出 Profile（eUICC 卡）

# 方法2: 在命令中直接设置
LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900000100" lpac chip info

# 其他命令
lpac driver apdu list   # 列出 APDU 驱动
lpac version            # 查看版本
```

**使用 EasyLPAC 应用程序（最推荐，最简单）：**

1. 打开 EasyLPAC 应用程序
2. 在"设置"标签页中，点击"测试 AID"按钮
3. 应用程序会自动测试 `aid.txt` 中的所有 AID，找到能成功读取卡片的 AID
4. 找到后会自动设置并显示成功消息

**注意：** 
- macOS 上 Homebrew 没有 `pcsc-tools` 包
- `opensc-tool` 可能无法识别某些卡片类型（如 eUICC），显示 "Unresponsive card"
- `lpac` 工具使用 PCSC 直接通信，更适合 eUICC 卡和 SIM 卡
- **如果 `lpac chip info` 返回 `code:-1` 和 `euicc_init`，说明 AID 不正确**
- **如果看到 `Sharing violation` 错误，说明读卡器被占用：**
  - 关闭 EasyLPAC 应用程序（如果正在运行）
  - 关闭其他使用读卡器的程序
  - 等待几秒后重试
- **最推荐使用 EasyLPAC 应用程序的"测试 AID"功能，可以自动查找正确的 AID**

### 5. 合并多个AID文件

如果有多个AID文件需要合并：

```bash
# 使用 fetch_aids.go 工具
go run fetch_aids.go

# 或者手动合并
cat aid1.txt aid2.txt aid3.txt | sort | uniq > aid_merged.txt
```

## AID格式说明

### 格式要求
- **长度**: 4-32个十六进制字符（2-16字节）
- **格式**: 必须是有效的十六进制字符串（0-9, A-F）
- **长度**: 必须是偶数（每个字节用2个十六进制字符表示）

### 常见AID前缀
- `A000000559`: eUICC相关（ISD-R）
- `A000000087`: 3GPP USIM应用
- `A000000003`: Visa支付
- `A000000004`: MasterCard支付
- `A000000025`: American Express
- `A000000065`: JCB支付

## 注意事项

1. **去重**: 合并AID时，相同AID只保留一个描述（保留更详细的）
2. **验证**: 添加AID前，确保格式正确（有效的十六进制字符串）
3. **备份**: 更新前建议备份现有 `aid.txt` 文件
4. **测试**: 添加新AID后，使用"测试AID"功能验证是否有效

## 贡献

如果您有新的AID数据源或发现更多AID，欢迎：
1. 直接编辑 `aid.txt` 文件添加
2. 或提交Issue/PR分享AID数据源

