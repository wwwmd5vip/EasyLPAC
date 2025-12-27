# STK (SIM Tool Kit) 程序开发指南

## 概述

STK (SIM Tool Kit) 是 GSM 系统的一个标准，允许 SIM 卡发起各类增值服务操作。开发并写入 STK 程序到 SIM 卡是一个复杂的过程，需要专门的工具和权限。

## 读取现有 STK 程序

在开发新的 STK 程序之前，您可能想要先读取 SIM 卡上现有的 STK 程序代码。这可以帮助您：
- 了解现有 STK 程序的结构
- 学习 STK 程序的实现方式
- 分析已安装的应用

### 方法 1：使用 Python 脚本（推荐）

已提供 `read_stk_applet.py` 脚本，可以：
- 枚举卡片上的所有应用程序
- 读取应用程序的初始响应数据
- 读取应用程序相关的文件
- 使用 GET DATA 命令获取应用信息

**使用方法：**
```bash
# 安装依赖
pip install pyscard

# 运行脚本
python3 read_stk_applet.py
```

### 方法 2：使用 OpenSC 工具

已提供 `read_stk_opensc.sh` 脚本，使用 OpenSC 工具读取：

```bash
# 运行脚本
./read_stk_opensc.sh
```

**功能：**
- 列出卡片上的文件
- 读取应用目录文件 (EF_DIR)
- 尝试选择并读取 STK 应用
- 使用 pkcs15-tool 列出应用

### 方法 3：使用 GlobalPlatform Pro

已提供 `read_stk_gp.sh` 脚本，使用 GlobalPlatform Pro 工具读取：

```bash
# 运行脚本（自动使用项目中的 gp.jar）
./read_stk_gp.sh
```

**手动使用 GlobalPlatform Pro：**

```bash
# 列出卡片上的应用
java -jar gp.jar -l

# 读取应用的详细信息
java -jar gp.jar --info A0000005591010FFFFFFFF8900000177

# 导出应用（如果支持）
java -jar gp.jar --get-data A0000005591010FFFFFFFF8900000177

# 列出应用文件
java -jar gp.jar --list-files A0000005591010FFFFFFFF8900000177
```

**常用命令：**
- `java -jar gp.jar -l` - 列出所有应用
- `java -jar gp.jar --info <AID>` - 获取应用信息
- `java -jar gp.jar --get-data <AID>` - 获取应用数据
- `java -jar gp.jar --list-files <AID>` - 列出应用文件
- `java -jar gp.jar --delete <AID>` - 删除应用（需要 ADM 密钥）
- `java -jar gp.jar --install <CAP文件>` - 安装应用（需要 ADM 密钥）

### 方法 4：使用 EasyLPAC/lpac

虽然 EasyLPAC/lpac 主要用于 eUICC，但可以用于：
- 读取卡片基本信息
- 测试和验证 AID
- 确认应用是否存在

```bash
# 测试 STK AID
LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900000177" ./lpac chip info
```

### 注意事项

1. **权限限制**：读取应用数据可能需要认证，某些数据可能受保护
2. **应用格式**：已安装的应用通常是编译后的 CAP 文件，不是源代码
3. **逆向工程**：从 CAP 文件还原源代码需要专门的工具和知识
4. **法律合规**：确保在合法和安全的前提下进行操作

## 重要说明

### EasyLPAC/lpac 的局限性

**EasyLPAC 和 lpac 工具主要用于 eUICC 卡的管理，不支持直接编写 STK 程序到传统 SIM 卡。**

- **EasyLPAC/lpac 的功能**：
  - 管理 eUICC 卡的 Profile（下载、安装、删除、启用、禁用）
  - 读取卡片信息
  - 管理通知
  - **不支持**：编写 STK 程序到传统 SIM 卡

- **传统 SIM 卡的 STK 编程**：
  - 需要专门的 SIM 卡编程工具
  - 需要运营商或卡片制造商的权限和密钥（如 ADM 密钥）
  - 需要 STK 应用程序的二进制文件（通常是 CAP 文件）

## STK 程序开发流程

### 1. 开发 STK 应用程序

STK 应用程序通常使用 **Java Card** 技术开发：

#### 开发工具
- **Java Card Development Kit (JDK)**
- **Java Card API**
- **Java Card 开发环境**（如 Eclipse 插件）

#### 开发步骤
1. **编写 Java Card Applet**
   ```java
   package com.example.stk;
   
   import javacard.framework.*;
   import javacardx.apdu.*;
   
   public class StkApplet extends Applet {
       public static void install(byte[] bArray, short bOffset, byte bLength) {
           new StkApplet().register();
       }
       
       public void process(APDU apdu) {
           // STK 命令处理逻辑
       }
   }
   ```

2. **编译为 CAP 文件**
   - 使用 Java Card 编译器将 Java 代码编译为 CAP（Converted Applet）文件
   - CAP 文件是 Java Card 应用程序的二进制格式

3. **测试和调试**
   - 使用 Java Card 模拟器进行测试
   - 在真实卡片上测试（需要开发卡）

### 2. 写入 STK 到 SIM 卡

#### 方法 1：使用 SIM 卡编程工具（需要权限和密钥）

**工具示例：**
- **SIMalliance Toolbox**：SIMalliance 提供的工具集
- **CardOS Toolbox**：Siemens CardOS 工具
- **JCOP Tools**：NXP JCOP 工具
- **运营商专用工具**：各运营商提供的工具

**要求：**
- **ADM 密钥**：需要卡片的管理密钥（Administrative Key）
- **权限**：通常只有运营商或卡片制造商有权限
- **开发卡**：使用可编程的开发卡进行测试

**基本流程：**
```bash
# 1. 连接到读卡器
# 2. 认证（使用 ADM 密钥）
# 3. 选择 MF (Master File)
# 4. 创建应用目录（如果需要）
# 5. 安装 CAP 文件
# 6. 配置 AID
# 7. 激活应用
```

#### 方法 2：使用 OpenSC 工具（有限支持）

```bash
# 安装 OpenSC
brew install opensc  # macOS
sudo apt-get install opensc  # Linux

# 使用 pkcs15-tool 管理应用（需要权限）
pkcs15-tool --list-applications
pkcs15-tool --install-applet --aid A0000000871004FF86FF4989 --cap-file applet.cap
```

**注意：** 大多数商业 SIM 卡不允许未授权安装应用，需要 ADM 密钥。

#### 方法 3：使用 Java Card 开发工具

**工具：**
- **GlobalPlatform Pro (gp)**：开源工具，支持 Java Card 应用管理
- **JCardSim**：Java Card 模拟器
- **JCIDE**：Java Card 集成开发环境

**示例（使用 GlobalPlatform）：**
```bash
# 安装 GlobalPlatform Pro
# 下载：https://github.com/martinpaljak/GlobalPlatformPro

# 列出卡片上的应用
gp -l

# 安装 CAP 文件
gp --install applet.cap --key 404142434445464748494A4B4C4D4E4F

# 删除应用
gp --delete A0000000871004FF86FF4989 --key 404142434445464748494A4B4C4D4E4F
```

### 3. 使用开发卡

**推荐使用开发卡进行 STK 开发：**

1. **购买开发卡**
   - 可编程的 Java Card 开发卡
   - 通常带有默认的 ADM 密钥
   - 允许安装和删除应用

2. **开发卡供应商**
   - NXP JCOP 开发卡
   - Infineon 开发卡
   - 其他 Java Card 开发卡

3. **优势**
   - 不需要运营商权限
   - 可以自由安装和删除应用
   - 适合开发和测试

## STK 程序结构

### AID 配置

STK 应用程序需要配置正确的 AID：

```java
// 在 applet 中定义 AID
private static final byte[] AID_STK = {
    (byte)0xA0, (byte)0x00, (byte)0x00, (byte)0x00, 
    (byte)0x87, (byte)0x10, (byte)0x04, (byte)0xFF, 
    (byte)0x86, (byte)0xFF, (byte)0x49, (byte)0x89
};

public static void install(byte[] bArray, short bOffset, byte bLength) {
    new StkApplet().register(AID_STK, (short)0, (byte)AID_STK.length);
}
```

### STK 命令处理

STK 应用程序需要处理标准的 STK 命令：

- **ENVELOPE**：从手机接收命令
- **TERMINAL RESPONSE**：向手机发送响应
- **FETCH**：从卡片获取待发送的命令
- **PROACTIVE COMMAND**：卡片主动发送命令

## 开发资源

### 文档和规范
- **ETSI TS 102 223**：SIM Tool Kit 规范
- **ETSI TS 102 241**：UICC Application Programming Interface
- **Java Card Platform Specification**：Java Card 平台规范
- **GlobalPlatform Card Specification**：GlobalPlatform 卡片规范

### 开发工具
- **Java Card Development Kit**：Oracle 官方 JDK
- **Eclipse Java Card Plugin**：Eclipse 插件
- **GlobalPlatform Pro**：开源工具
- **JCardSim**：Java Card 模拟器

### 参考实现
- **Android STK 应用**：[AOSP Stk.git](https://android.googlesource.com/platform/packages/apps/Stk.git)
- **Java Card 示例**：Oracle 官方示例

## 安全注意事项

1. **权限要求**
   - 写入 STK 到商业 SIM 卡需要运营商或卡片制造商的权限
   - 需要 ADM 密钥或其他管理密钥

2. **法律合规**
   - 确保在合法和安全的前提下进行操作
   - 不要修改他人的 SIM 卡
   - 遵守相关法律法规

3. **开发建议**
   - 使用开发卡进行开发和测试
   - 不要在生产卡上直接测试
   - 备份重要数据

## 使用 EasyLPAC 的场景

虽然 EasyLPAC 不能直接编写 STK 到传统 SIM 卡，但可以用于：

1. **读取卡片信息**
   - 使用"测试 AID"功能找到正确的 AID
   - 读取卡片基本信息

2. **eUICC 卡管理**
   - 如果是 eUICC 卡，可以管理 Profile
   - 下载和安装 Profile（可能包含 STK 功能）

3. **验证 AID**
   - 测试 STK AID 是否正确
   - 验证卡片是否支持 STK

## 总结

编写 STK 程序到 SIM 卡是一个复杂的过程，需要：

1. **开发工具**：Java Card 开发环境
2. **编程工具**：SIM 卡编程工具（如 GlobalPlatform Pro）
3. **权限和密钥**：ADM 密钥或其他管理密钥
4. **开发卡**：推荐使用可编程的开发卡

**EasyLPAC 主要用于 eUICC 卡的管理，不支持直接编写 STK 到传统 SIM 卡。**

如果您需要开发 STK 程序，建议：
1. 使用 Java Card 开发工具创建 STK 应用
2. 使用开发卡进行测试
3. 使用 GlobalPlatform Pro 或其他工具安装到卡片
4. 使用 EasyLPAC 验证 AID 和读取卡片信息

