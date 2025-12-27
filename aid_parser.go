package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadAidList 从aid.txt文件加载AID列表
// 返回解析的AID列表和错误信息
func LoadAidList() ([]*AidItem, error) {
	var aidList []*AidItem
	
	// 尝试多个可能的aid.txt文件路径
	var aidFilePath string
	var file *os.File
	var err error
	
	// 1. 尝试可执行文件所在目录
	exePath, err := os.Executable()
	if err == nil {
		exePath, err = filepath.EvalSymlinks(exePath)
		if err == nil {
			exeDir := filepath.Dir(exePath)
			aidFilePath = filepath.Join(exeDir, "aid.txt")
			file, err = os.Open(aidFilePath)
			if err == nil {
				defer file.Close()
			}
		}
	}
	
	// 2. 如果找不到，尝试当前工作目录
	if file == nil {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			aidFilePath = filepath.Join(wd, "aid.txt")
			file, err = os.Open(aidFilePath)
			if err == nil {
				defer file.Close()
			}
		}
	}
	
	// 3. 如果还是找不到，返回错误
	if file == nil {
		return nil, fmt.Errorf("aid.txt not found in executable directory or current working directory")
	}
	
	// 读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Filetype:") || strings.HasPrefix(line, "Version:") {
			continue
		}
		
		// 解析格式：AID: 描述
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		aid := strings.TrimSpace(parts[0])
		description := strings.TrimSpace(parts[1])
		
		// 清理AID：去除空格，转为大写
		aidHex := strings.ToUpper(strings.ReplaceAll(aid, " ", ""))
		
		// 验证是否为有效的十六进制字符串
		if !isValidHex(aidHex) {
			continue
		}
		
		// AID长度必须是偶数（因为每个字节用2个十六进制字符表示）
		if len(aidHex)%2 != 0 {
			continue
		}
		
		// 只处理长度在4-32之间的AID（2-16字节）
		// 注意：AID长度不固定，根据ISO 7816标准，长度范围是5-16字节（10-32个十六进制字符）
		if len(aidHex) < 4 || len(aidHex) > 32 {
			continue
		}
		
		// 判断是否为eUICC相关AID（以A000000559开头）
		// eUICC的ISD-R AID通常是32位，但基础标识可能是10位
		isEuicc := strings.HasPrefix(aidHex, "A000000559")
		
		// 保存原始AID（不补齐）
		// 注意：不要补齐AID，因为：
		// 1. AID长度不固定，补齐会丢失原始信息
		// 2. 补齐后的AID可能不是有效的AID
		// 3. lpac工具需要原始长度的AID
		aidItem := &AidItem{
			AID:         aidHex,
			Description: description,
			IsEuicc:     isEuicc,
		}
		
		aidList = append(aidList, aidItem)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return aidList, nil
}

// isValidHex 检查字符串是否为有效的十六进制字符串
func isValidHex(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// FindBestAid 根据当前配置的AID，从AID列表中查找最佳匹配
// 优先返回eUICC相关的AID，如果当前AID在列表中，则返回当前AID
func FindBestAid(currentAid string, aidList []*AidItem) *AidItem {
	currentAid = strings.ToUpper(strings.ReplaceAll(currentAid, " ", ""))
	
	// 首先尝试查找完全匹配的AID（精确匹配）
	for _, item := range aidList {
		if item.AID == currentAid {
			return item
		}
	}
	
	// 如果没有完全匹配，尝试前缀匹配（例如：当前AID是32位，列表中是10位前缀）
	// 例如：当前AID是 A0000005591010FFFFFFFF8900000100，列表中有 A000000559
	for _, item := range aidList {
		if strings.HasPrefix(currentAid, item.AID) || strings.HasPrefix(item.AID, currentAid) {
			return item
		}
	}
	
	// 如果没有匹配，返回第一个eUICC相关的AID（优先32位的ISD-R AID）
	// 优先返回32位的eUICC AID
	for _, item := range aidList {
		if item.IsEuicc && len(item.AID) == 32 {
			return item
		}
	}
	// 如果没有32位的，返回其他eUICC相关的AID
	for _, item := range aidList {
		if item.IsEuicc {
			return item
		}
	}
	
	// 如果都没有，返回列表中的第一个
	if len(aidList) > 0 {
		return aidList[0]
	}
	
	return nil
}

// SearchAidList 在AID列表中搜索匹配的AID
// 支持按AID值或描述搜索
func SearchAidList(query string, aidList []*AidItem) []*AidItem {
	if query == "" {
		return aidList
	}
	
	query = strings.ToUpper(strings.TrimSpace(query))
	var results []*AidItem
	
	for _, item := range aidList {
		// 搜索AID值
		if strings.Contains(strings.ToUpper(item.AID), query) {
			results = append(results, item)
			continue
		}
		
		// 搜索描述
		if strings.Contains(strings.ToUpper(item.Description), query) {
			results = append(results, item)
			continue
		}
	}
	
	return results
}

// TestAid 测试指定的AID是否能成功读取卡片信息
// 返回true表示AID有效，false表示无效
// 支持eUICC卡和传统SIM卡（如移动、联通、电信的手机卡）
// 使用轻量级测试方法提高效率：优先使用 profile list（eUICC），失败则使用 chip info（通用）
func TestAid(aid string) bool {
	// 保存当前AID
	originalAid := ConfigInstance.LpacAID
	
	// 临时设置测试AID
	ConfigInstance.LpacAID = aid
	defer func() {
		// 恢复原始AID
		ConfigInstance.LpacAID = originalAid
	}()
	
	// 对于eUICC卡，优先使用更轻量的 profile list 命令来测试AID
	// 这比 chip info 更快，因为不需要读取完整的芯片信息
	_, err := LpacProfileList()
	if err == nil {
		// profile list 成功，说明AID有效（eUICC卡）
		return true
	}
	
	// 如果 profile list 失败，尝试使用 chip info
	// 注意：chip info 可以用于：
	// 1. eUICC卡（如果profile list失败但chip info成功）
	// 2. 传统SIM卡（如移动、联通、电信的手机卡）
	//    传统SIM卡不支持profile list，但chip info可能能够读取基本信息
	// 注意：runLpac 已经内置了5秒超时，所以这里不需要额外超时
	_, err = LpacChipInfo()
	return err == nil
}

// FindWorkingAid 自动测试AID列表，找到第一个能成功读取卡片的AID
// 优先测试eUICC相关的AID，然后测试其他AID
// 返回找到的有效AID，如果没有找到则返回nil
func FindWorkingAid(aidList []*AidItem) *AidItem {
	// 首先测试eUICC相关的AID
	var euiccAids []*AidItem
	var otherAids []*AidItem
	
	for _, item := range aidList {
		if item.IsEuicc {
			euiccAids = append(euiccAids, item)
		} else {
			otherAids = append(otherAids, item)
		}
	}
	
	// 先测试eUICC相关的AID
	for _, item := range euiccAids {
		if TestAid(item.AID) {
			return item
		}
	}
	
	// 如果eUICC相关的AID都不行，测试其他AID
	for _, item := range otherAids {
		if TestAid(item.AID) {
			return item
		}
	}
	
	return nil
}


