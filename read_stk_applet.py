#!/usr/bin/env python3
"""
从 SIM 卡读取 STK 应用程序代码的脚本
需要安装: pip install pyscard
"""

from smartcard.System import readers
from smartcard.util import toHexString, toBytes, toASCIIString
from smartcard.ATR import ATR
from smartcard.CardConnection import CardConnection
import sys

def select_application(connection, aid_hex):
    """选择指定的应用程序"""
    aid_bytes = toBytes(aid_hex)
    select_cmd = [0x00, 0xA4, 0x04, 0x00, len(aid_bytes)] + aid_bytes
    data, sw1, sw2 = connection.transmit(select_cmd)
    return sw1 == 0x90 and sw2 == 0x00, data, sw1, sw2

def read_binary(connection, offset, length):
    """读取二进制数据"""
    p1 = (offset >> 8) & 0xFF
    p2 = offset & 0xFF
    read_cmd = [0x00, 0xB0, p1, p2, length]
    data, sw1, sw2 = connection.transmit(read_cmd)
    if sw1 == 0x90 and sw2 == 0x00:
        return True, data
    return False, None

def get_response(connection, expected_length):
    """获取响应数据"""
    get_resp_cmd = [0x00, 0xC0, 0x00, 0x00, expected_length]
    data, sw1, sw2 = connection.transmit(get_resp_cmd)
    if sw1 == 0x90 and sw2 == 0x00:
        return True, data
    return False, None

def read_file(connection, file_id):
    """读取文件（使用文件 ID）"""
    # SELECT FILE by ID
    p1 = (file_id >> 8) & 0xFF
    p2 = file_id & 0xFF
    select_cmd = [0x00, 0xA4, 0x00, 0x00, 0x02, p1, p2]
    data, sw1, sw2 = connection.transmit(select_cmd)
    if sw1 != 0x90 or sw2 != 0x00:
        return False, None
    
    # READ BINARY
    read_cmd = [0x00, 0xB0, 0x00, 0x00, 0x00]  # 读取长度由响应决定
    data, sw1, sw2 = connection.transmit(read_cmd)
    if sw1 == 0x6C:  # 长度错误，使用返回的长度
        read_cmd = [0x00, 0xB0, 0x00, 0x00, sw2]
        data, sw1, sw2 = connection.transmit(read_cmd)
    
    if sw1 == 0x90 and sw2 == 0x00:
        return True, data
    return False, None

def list_applications(connection):
    """尝试列出卡片上的应用程序"""
    print("\n=== 尝试枚举应用程序 ===")
    
    # 方法1: 尝试读取 EF_DIR (2F00) - 应用目录文件
    print("\n1. 尝试读取 EF_DIR (应用目录文件)...")
    success, data = read_file(connection, 0x2F00)
    if success and data:
        print(f"  EF_DIR 内容: {toHexString(data)}")
        # 解析 AID
        i = 0
        while i < len(data):
            if i + 1 < len(data):
                tag = data[i]
                length = data[i + 1]
                if tag == 0x61 and i + 1 + length <= len(data):
                    aid_data = data[i + 2:i + 2 + length]
                    if len(aid_data) >= 5:
                        aid = toHexString(aid_data[:5])
                        print(f"  找到 AID: {aid}")
                i += 2 + length
            else:
                break
    else:
        print("  无法读取 EF_DIR")
    
    # 方法2: 尝试常见的 AID
    print("\n2. 测试常见 AID...")
    common_aids = [
        "A0000005591010FFFFFFFF8900000177",  # eUICC
        "A0000000871004FF86FF4989",  # ISIM/STK
        "A0000000871002FF86FF4989",  # USIM/STK
        "A0000000871002FF49FF0589",  # USIM
        "A000000087",  # 3GPP 基础
        "A0000000030000",  # SIM
    ]
    
    found_aids = []
    for aid_hex in common_aids:
        success, data, sw1, sw2 = select_application(connection, aid_hex)
        if success:
            print(f"  ✓ {aid_hex}: 找到应用")
            print(f"    响应: {toHexString(data)}")
            found_aids.append((aid_hex, data))
        elif sw1 == 0x6A and sw2 == 0x82:
            print(f"  ✗ {aid_hex}: 应用不存在")
        else:
            print(f"  ? {aid_hex}: 状态 {hex(sw1)}{hex(sw2)}")
    
    return found_aids

def read_application_data(connection, aid_hex):
    """读取应用程序的数据"""
    print(f"\n=== 读取应用程序数据: {aid_hex} ===")
    
    # 选择应用
    success, data, sw1, sw2 = select_application(connection, aid_hex)
    if not success:
        print(f"无法选择应用: {sw1:02X}{sw2:02X}")
        return None
    
    print(f"应用选择成功")
    print(f"初始响应: {toHexString(data)}")
    
    # 尝试读取应用数据
    app_data = {
        'aid': aid_hex,
        'initial_response': toHexString(data),
        'files': []
    }
    
    # 尝试读取常见的文件（包括STK菜单相关文件）
    print("\n尝试读取应用文件...")
    common_files = [
        (0x6F07, "EF_IMSI"),
        (0x6F20, "EF_SST"),
        (0x6FAD, "EF_ADN"),
        (0x6F3A, "EF_AD"),
        (0x6F3B, "EF_MSISDN"),
        # STK 菜单相关文件
        (0x6F30, "EF_MENU"),          # STK菜单文件
        (0x6F10, "EF_PROACTIVE"),      # Proactive命令
        (0x6F38, "EF_UST"),            # USIM Service Table
        (0x6F56, "EF_EST"),            # Enabled Services Table
        (0x6F06, "EF_ARR"),            # Access Rule Reference
        (0x6F42, "EF_MSK"),            # Menu Selection Key
        (0x6F43, "EF_MSK_EXT"),        # Extended Menu Selection Key
    ]
    
    for file_id, file_name in common_files:
        success, file_data = read_file(connection, file_id)
        if success and file_data:
            print(f"  ✓ {file_name} (0x{file_id:04X}): {toHexString(file_data[:50])}...")
            app_data['files'].append({
                'id': file_id,
                'name': file_name,
                'data': toHexString(file_data),
                'raw_data': file_data
            })
        else:
            print(f"  ✗ {file_name} (0x{file_id:04X}): 无法读取")
    
    return app_data

def parse_stk_menu(menu_data):
    """解析STK菜单文件数据"""
    if not menu_data or len(menu_data) < 2:
        print("  菜单数据为空或格式不正确")
        return
    
    print("\n  菜单结构解析:")
    idx = 0
    
    try:
        # STK菜单文件格式（根据ETSI TS 102 223）：
        # 第一个字节通常是菜单项数量或菜单标识符
        
        # 尝试解析菜单项数量
        if len(menu_data) > 0:
            num_items = menu_data[0]
            print(f"  菜单项数量/标识符: 0x{num_items:02X} ({num_items})")
            idx = 1
            
            # 解析每个菜单项
            item_idx = 0
            while idx < len(menu_data) and item_idx < num_items and num_items < 20:  # 限制最大20项
                if idx + 1 >= len(menu_data):
                    break
                
                # 菜单项结构（简化版，根据ETSI TS 102 223）：
                # - 1字节：菜单项标识符
                # - 1字节：菜单项类型/标志
                # - 1字节：菜单项名称长度
                # - N字节：菜单项名称（UTF-8或GSM 7-bit）
                # - 可选：图标标识符等
                
                menu_id = menu_data[idx]
                idx += 1
                
                if idx >= len(menu_data):
                    break
                
                menu_type = menu_data[idx]
                idx += 1
                
                if idx >= len(menu_data):
                    break
                
                name_len = menu_data[idx]
                idx += 1
                
                if name_len > 0 and idx + name_len <= len(menu_data):
                    name_bytes = menu_data[idx:idx+name_len]
                    # 尝试解码为UTF-8或ASCII
                    try:
                        # 尝试UTF-8解码
                        menu_name = name_bytes.decode('utf-8', errors='replace')
                    except:
                        try:
                            # 尝试ASCII解码
                            menu_name = ''.join([chr(b) if 32 <= b < 127 else f'\\x{b:02X}' for b in name_bytes])
                        except:
                            menu_name = toHexString(name_bytes)
                    
                    idx += name_len
                    
                    print(f"    菜单项 {item_idx + 1}:")
                    print(f"      ID: 0x{menu_id:02X} ({menu_id})")
                    print(f"      类型: 0x{menu_type:02X}")
                    print(f"      名称长度: {name_len}")
                    print(f"      名称: {menu_name}")
                    print(f"      名称(HEX): {toHexString(name_bytes)}")
                    
                    item_idx += 1
                else:
                    # 如果名称长度为0或超出范围，可能不是标准格式
                    break
        
        # 如果无法按标准格式解析，显示原始数据
        if idx <= 1 or item_idx == 0:  # 说明没有找到菜单项
            print("  无法按标准格式解析，显示原始数据:")
            print(f"  完整数据: {toHexString(menu_data)}")
            print(f"  数据长度: {len(menu_data)} 字节")
            
            # 尝试查找可能的文本字符串
            print("\n  可能的文本内容:")
            text_chars = []
            text_start = 0
            for i, byte in enumerate(menu_data):
                if 32 <= byte < 127:  # 可打印ASCII字符
                    if not text_chars:
                        text_start = i
                    text_chars.append(chr(byte))
                elif text_chars:
                    if len(text_chars) >= 3:  # 至少3个连续字符才显示
                        print(f"    位置 {text_start}-{i-1}: {''.join(text_chars)}")
                    text_chars = []
            if text_chars and len(text_chars) >= 3:
                print(f"    位置 {text_start}-{len(menu_data)-1}: {''.join(text_chars)}")
    
    except Exception as e:
        print(f"  解析错误: {e}")
        print(f"  原始数据: {toHexString(menu_data)}")

def main():
    """主函数"""
    print("=== SIM 卡 STK 应用程序读取工具 ===")
    print("")
    
    # 获取读卡器列表
    r = readers()
    if not r:
        print("错误: 未找到读卡器")
        return
    
    print(f"找到 {len(r)} 个读卡器")
    
    # 使用第一个读卡器
    reader = r[0]
    print(f"使用读卡器: {reader}")
    
    try:
        # 创建连接
        connection = reader.createConnection()
        
        # 尝试不同的连接方式
        print("\n尝试连接卡片...")
        connected = False
        last_error = None
        
        # 方法1: 尝试自动协议（默认，最常用）
        try:
            connection.connect()
            connected = True
            print("  ✓ 使用自动协议连接成功")
        except Exception as e:
            last_error = e
            print(f"  ✗ 自动协议连接失败: {e}")
        
        # 方法2: 如果自动协议失败，尝试 T=0 协议
        if not connected:
            try:
                connection.connect(CardConnection.T0_protocol)
                connected = True
                print("  ✓ 使用 T=0 协议连接成功")
            except Exception as e:
                last_error = e
                print(f"  ✗ T=0 协议连接失败: {e}")
        
        # 方法3: 如果 T=0 失败，尝试 T=1 协议
        if not connected:
            try:
                connection.connect(CardConnection.T1_protocol)
                connected = True
                print("  ✓ 使用 T=1 协议连接成功")
            except Exception as e:
                last_error = e
                print(f"  ✗ T=1 协议连接失败: {e}")
        
        if not connected:
            print("\n" + "="*60)
            print("错误: 无法连接到卡片")
            print("="*60)
            print("\n可能的原因：")
            print("  1. 读卡器被其他程序占用（如 EasyLPAC 正在运行）")
            print("  2. 卡片未正确插入或接触不良")
            print("  3. 读卡器驱动问题")
            print("  4. 卡片需要先选择正确的 AID 才能连接")
            print("  5. PCSC 服务未正确启动")
            print("\n解决方案：")
            print("  1. 关闭 EasyLPAC 应用程序（如果正在运行）")
            print("  2. 等待 3-5 秒让 PCSC 服务释放资源")
            print("  3. 检查卡片是否正确插入，尝试重新插入卡片")
            print("  4. 尝试使用 lpac 工具先建立连接：")
            print("     ./lpac driver apdu list")
            print("     ./read_stk_with_lpac.sh")
            print("  5. 尝试使用 EasyLPAC 应用程序的'测试 AID'功能")
            print("  6. 重启 PCSC 服务（macOS）：")
            print("     sudo launchctl stop org.opensc.pcscd")
            print("     sudo launchctl start org.opensc.pcscd")
            print("\n替代方案：")
            print("  - 使用 read_stk_with_lpac.sh 脚本（使用 lpac 工具）")
            print("  - 使用 read_stk_opensc.sh 脚本（使用 OpenSC 工具）")
            print("  - 使用 EasyLPAC 应用程序的'测试 AID'功能")
            if last_error:
                print(f"\n最后错误: {last_error}")
            return
        
        print("\n卡片信息:")
        atr = connection.getATR()
        print(f"  ATR: {toHexString(atr)}")
        
        # 解析 ATR
        try:
            atr_obj = ATR(atr)
            print(f"  历史字节: {toHexString(atr_obj.getHistoricalBytes())}")
        except:
            pass
        
        # 列出应用程序
        found_aids = list_applications(connection)
        
        # 读取每个找到的应用的数据
        if found_aids:
            print("\n" + "="*50)
            print("读取应用程序数据")
            print("="*50)
            
            for aid_hex, initial_data in found_aids:
                app_data = read_application_data(connection, aid_hex)
                if app_data:
                    print(f"\n应用程序摘要:")
                    print(f"  AID: {app_data['aid']}")
                    print(f"  初始响应: {app_data['initial_response']}")
                    print(f"  找到文件数: {len(app_data['files'])}")
                    
                    # 特别处理STK菜单文件和其他STK相关文件
                    stk_files = ['EF_MENU', 'EF_PROACTIVE', 'EF_UST', 'EF_EST', 'EF_ARR', 'EF_MSK', 'EF_MSK_EXT']
                    for file_info in app_data['files']:
                        if file_info['name'] in stk_files:
                            if file_info['name'] == 'EF_MENU':
                                print("\n" + "="*50)
                                print("STK 菜单文件 (EF_MENU)")
                                print("="*50)
                                print(f"  文件ID: 0x{file_info['id']:04X}")
                                print(f"  数据长度: {len(file_info['raw_data'])} 字节")
                                print(f"  原始数据: {file_info['data']}")
                                parse_stk_menu(file_info['raw_data'])
                            else:
                                print(f"\n{file_info['name']}:")
                                print(f"  文件ID: 0x{file_info['id']:04X}")
                                print(f"  数据长度: {len(file_info['raw_data'])} 字节")
                                print(f"  原始数据: {file_info['data']}")
        
        # 尝试使用 GET DATA 命令读取应用信息
        print("\n" + "="*50)
        print("尝试使用 GET DATA 命令")
        print("="*50)
        
        # GET DATA 命令示例
        get_data_tags = [
            (0x0042, "ICCID"),
            (0x0043, "Application Label"),
            (0x0045, "Application Template"),
        ]
        
        for tag, name in get_data_tags:
            p1 = (tag >> 8) & 0xFF
            p2 = tag & 0xFF
            get_data_cmd = [0x00, 0xCA, p1, p2, 0x00]
            data, sw1, sw2 = connection.transmit(get_data_cmd)
            if sw1 == 0x6C:
                # 使用返回的长度重试
                get_data_cmd = [0x00, 0xCA, p1, p2, sw2]
                data, sw1, sw2 = connection.transmit(get_data_cmd)
            
            if sw1 == 0x90 and sw2 == 0x00:
                print(f"  ✓ {name} (0x{tag:04X}): {toHexString(data)}")
            else:
                print(f"  ✗ {name} (0x{tag:04X}): 无法读取 ({sw1:02X}{sw2:02X})")
        
    except Exception as e:
        print(f"错误: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()

