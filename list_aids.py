#!/usr/bin/env python3
"""
从智能卡读取AID列表的Python脚本
需要安装: pip install pyscard
"""

from smartcard.System import readers
from smartcard.util import toHexString, toBytes
import sys

def list_aids():
    """列出智能卡上的所有AID"""
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
        connection.connect()
        
        print("\n卡片信息:")
        print(f"  ATR: {toHexString(connection.getATR())}")
        
        # 尝试使用 SELECT 命令枚举应用
        # 注意：不同的卡片类型可能需要不同的命令
        print("\n尝试枚举应用...")
        
        # 方法1: SELECT by name (空AID，获取主应用)
        SELECT_MAIN = [0x00, 0xA4, 0x04, 0x00, 0x00]
        data, sw1, sw2 = connection.transmit(SELECT_MAIN)
        if sw1 == 0x90 and sw2 == 0x00:
            print(f"  主应用响应: {toHexString(data)}")
        
        # 方法2: 尝试常见的AID前缀
        common_aids = [
            "A000000559",  # eUICC
            "A000000087",  # USIM
            "A000000003",  # Visa
            "A000000004",  # MasterCard
        ]
        
        print("\n测试常见AID:")
        for aid_prefix in common_aids:
            aid_bytes = toBytes(aid_prefix)
            SELECT = [0x00, 0xA4, 0x04, 0x00, len(aid_bytes)] + aid_bytes
            data, sw1, sw2 = connection.transmit(SELECT)
            if sw1 == 0x90 and sw2 == 0x00:
                print(f"  ✓ {aid_prefix}: 找到应用")
                if data:
                    print(f"    响应: {toHexString(data)}")
            elif sw1 == 0x6A and sw2 == 0x82:
                print(f"  ✗ {aid_prefix}: 应用不存在")
            else:
                print(f"  ? {aid_prefix}: 状态 {hex(sw1)}{hex(sw2)}")
        
        # 方法3: 尝试读取文件系统（如果支持）
        print("\n尝试读取文件系统...")
        # 这需要根据具体的卡片类型来实现
        
    except Exception as e:
        print(f"错误: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    list_aids()

