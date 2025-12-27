# 故障排除指南

## 常见错误和解决方案

### 1. "Sharing violation" 错误

**错误信息：**
```
SCardConnect() failed: 8010000B (Sharing violation.)
```

**原因：**
读卡器被另一个程序占用。通常是因为：
- EasyLPAC 应用程序正在运行
- 其他使用读卡器的程序正在运行
- 之前的程序没有正确释放读卡器资源

**解决方案：**

1. **关闭 EasyLPAC 应用程序**
   - 如果 EasyLPAC 正在运行，完全退出应用程序
   - 在 macOS 上，确保从 Dock 中退出（右键点击图标 → 退出）

2. **关闭其他使用读卡器的程序**
   - 检查是否有其他程序正在使用读卡器
   - 关闭所有可能使用读卡器的应用程序

3. **等待几秒后重试**
   - PCSC 服务需要一些时间来释放资源
   - 等待 3-5 秒后再次尝试

4. **重启 PCSC 服务（如果问题持续）**
   ```bash
   # macOS
   sudo launchctl stop org.opensc.pcscd
   sudo launchctl start org.opensc.pcscd
   
   # Linux
   sudo systemctl restart pcscd
   ```

5. **使用 EasyLPAC 应用程序代替命令行**
   - 如果需要在 EasyLPAC 运行时测试 AID，直接使用应用程序内的"测试 AID"功能
   - 这是最可靠的方法，因为应用程序已经管理好了读卡器资源

### 2. "euicc_init" 错误

**错误信息：**
```json
{"type":"lpa","payload":{"code":-1,"message":"euicc_init","data":""}}
```

**原因：**
- AID 不正确
- 卡片不是 eUICC 卡
- 卡片类型不匹配

**解决方案：**

1. **使用 EasyLPAC 应用程序的"测试 AID"功能**
   - 这是最可靠的方法
   - 应用程序会自动测试所有 AID，找到正确的

2. **手动测试不同的 AID**
   ```bash
   # 测试不同的 AID
   LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900000100" ./lpac chip info
   LPAC_CUSTOM_ISD_R_AID="A0000005591010FFFFFFFF8900050500" ./lpac chip info
   LPAC_CUSTOM_ISD_R_AID="A0000005591010000000008900000300" ./lpac chip info
   ```

3. **检查 aid.txt 文件**
   - 查看 `aid.txt` 文件中的 AID 列表
   - 尝试不同的 AID

### 3. "Unresponsive card" 错误（OpenSC）

**错误信息：**
```
Failed to connect to card: Unresponsive card (correctly inserted?)
```

**原因：**
- OpenSC 无法识别卡片类型
- 卡片类型不被 OpenSC 支持（如某些 eUICC 卡）

**解决方案：**

1. **使用 lpac 工具代替 OpenSC**
   - `lpac` 使用 PCSC 直接通信，更适合 eUICC 卡
   - 使用 `./test_aid_lpac.sh` 脚本

2. **使用 EasyLPAC 应用程序**
   - 应用程序已经配置好了所有必要的驱动
   - 直接使用应用程序的"测试 AID"功能

### 4. 读卡器未检测到

**错误信息：**
```
No card reader found
```

**解决方案：**

1. **检查读卡器连接**
   - 确保读卡器已正确连接到电脑
   - 检查 USB 连接是否稳定

2. **检查驱动安装**
   - macOS: 某些读卡器需要安装特定驱动（如 ACR38U）
   - Linux: 确保 `pcscd` 服务正在运行

3. **检查 PCSC 服务**
   ```bash
   # macOS
   pcsc_scan
   
   # Linux
   sudo systemctl status pcscd
   ```

### 5. AID 测试很慢

**原因：**
- 测试每个 AID 需要时间（每个 AID 测试约 5 秒超时）
- 如果 AID 列表很大，测试所有 AID 需要很长时间

**解决方案：**

1. **使用 EasyLPAC 应用程序的"测试 AID"功能**
   - 应用程序会显示进度
   - 找到有效 AID 后会自动停止

2. **使用取消功能**
   - 如果测试时间太长，可以点击"取消"按钮
   - 然后手动尝试一些常见的 AID

3. **优先测试 eUICC AID**
   - 如果知道是 eUICC 卡，可以只测试以 `A000000559` 开头的 AID
   - 在 EasyLPAC 中勾选"只显示 eUICC 相关 AID"后测试

## 最佳实践

1. **使用 EasyLPAC 应用程序**
   - 应用程序已经处理了所有资源管理
   - 使用"测试 AID"功能是最可靠的方法

2. **确保只有一个程序使用读卡器**
   - 不要同时运行多个使用读卡器的程序
   - 关闭不需要的程序

3. **定期更新 aid.txt**
   - 使用 `fetch_aids.go` 工具更新 AID 列表
   - 更多 AID 意味着更高的成功率

4. **保存有效的 AID**
   - 找到有效的 AID 后，在 EasyLPAC 中保存
   - 下次使用时就不需要重新测试

## 获取帮助

如果以上方法都无法解决问题，可以：
1. 检查日志文件（在 EasyLPAC 中点击"打开日志"）
2. 查看 `README_AID.md` 获取更多信息
3. 提交 Issue 到 GitHub 仓库

