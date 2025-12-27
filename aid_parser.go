package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadAidList 从aid.txt文件加载AID列表
// 返回解析的AID列表和错误信息
func LoadAidList() ([]*AidItem, error) {
	var aidList []*AidItem
	
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)
	
	// 构建aid.txt文件路径
	aidFilePath := filepath.Join(exeDir, "aid.txt")
	
	// 打开文件
	file, err := os.Open(aidFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
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
		
		// 只处理32位十六进制AID（去除所有非十六进制字符后长度为32）
		aidHex := strings.ToUpper(strings.ReplaceAll(aid, " ", ""))
		if len(aidHex) != 32 {
			// 如果不是32位，尝试补齐或跳过
			// 对于较短的AID，可能需要补齐，但这里我们只处理32位的
			continue
		}
		
		// 验证是否为有效的十六进制字符串
		if !isValidHex(aidHex) {
			continue
		}
		
		// 判断是否为eUICC相关AID（以A000000559开头）
		isEuicc := strings.HasPrefix(aidHex, "A000000559")
		
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
	
	// 首先尝试查找完全匹配的AID
	for _, item := range aidList {
		if item.AID == currentAid {
			return item
		}
	}
	
	// 如果没有完全匹配，返回第一个eUICC相关的AID
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
func TestAid(aid string) bool {
	// 保存当前AID
	originalAid := ConfigInstance.LpacAID
	
	// 临时设置测试AID
	ConfigInstance.LpacAID = aid
	
	// 尝试读取芯片信息
	_, err := LpacChipInfo()
	
	// 恢复原始AID
	ConfigInstance.LpacAID = originalAid
	
	// 如果没有错误，说明AID有效
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

