// +build ignore

package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// fetchAidsFromURL 从URL获取AID列表
func fetchAidsFromURL(url string) ([]string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取URL失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}
	
	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	return lines, nil
}

// mergeAidFiles 合并多个AID文件，去重
func mergeAidFiles(outputPath string, inputPaths []string) error {
	aidMap := make(map[string]string) // AID -> Description
	
	// 读取所有输入文件
	for _, inputPath := range inputPaths {
		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("警告: 无法打开文件 %s: %v\n", inputPath, err)
			continue
		}
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			
			// 跳过空行和注释
			if line == "" || strings.HasPrefix(line, "#") || 
			   strings.HasPrefix(line, "Filetype:") || strings.HasPrefix(line, "Version:") {
				continue
			}
			
			// 解析格式：AID: 描述
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			
			aid := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(parts[0], " ", "")))
			description := strings.TrimSpace(parts[1])
			
			// 验证AID格式
			if len(aid) < 4 || len(aid) > 32 || len(aid)%2 != 0 {
				continue
			}
			
			// 如果AID已存在，保留更长的描述
			if existingDesc, exists := aidMap[aid]; exists {
				if len(description) > len(existingDesc) {
					aidMap[aid] = description
				}
			} else {
				aidMap[aid] = description
			}
		}
		
		file.Close()
		if err := scanner.Err(); err != nil {
			fmt.Printf("警告: 读取文件 %s 时出错: %v\n", inputPath, err)
		}
	}
	
	// 写入输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()
	
	// 按AID排序（简单字符串排序）
	var sortedAids []string
	for aid := range aidMap {
		sortedAids = append(sortedAids, aid)
	}
	
	// 简单的字符串排序
	for i := 0; i < len(sortedAids)-1; i++ {
		for j := i + 1; j < len(sortedAids); j++ {
			if sortedAids[i] > sortedAids[j] {
				sortedAids[i], sortedAids[j] = sortedAids[j], sortedAids[i]
			}
		}
	}
	
	// 写入文件
	for _, aid := range sortedAids {
		_, err := fmt.Fprintf(outputFile, "%s: %s\n", aid, aidMap[aid])
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}
	
	fmt.Printf("成功合并 %d 个AID到 %s\n", len(aidMap), outputPath)
	return nil
}

// updateAidFile 更新aid.txt文件，从多个源获取AID
func updateAidFile() error {
	// AID数据源URL列表（示例）
	aidSources := []string{
		// 可以添加更多AID数据源URL
		// "https://example.com/aid-list.txt",
	}
	
	// 获取当前aid.txt路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("解析符号链接失败: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	aidFilePath := filepath.Join(exeDir, "aid.txt")
	
	// 备份现有文件
	backupPath := aidFilePath + ".backup." + time.Now().Format("20060102_150405")
	if _, err := os.Stat(aidFilePath); err == nil {
		source, err := os.Open(aidFilePath)
		if err == nil {
			defer source.Close()
			dest, err := os.Create(backupPath)
			if err == nil {
				io.Copy(dest, source)
				dest.Close()
				fmt.Printf("已备份现有aid.txt到: %s\n", backupPath)
			}
		}
	}
	
	// 从网络获取AID（如果有URL）
	var tempFiles []string
	for i, url := range aidSources {
		fmt.Printf("正在从 %s 获取AID...\n", url)
		lines, err := fetchAidsFromURL(url)
		if err != nil {
			fmt.Printf("警告: 从 %s 获取AID失败: %v\n", url, err)
			continue
		}
		
		// 保存到临时文件
		tempFile := filepath.Join(exeDir, fmt.Sprintf("aid_temp_%d.txt", i))
		tempFiles = append(tempFiles, tempFile)
		
		file, err := os.Create(tempFile)
		if err != nil {
			fmt.Printf("警告: 创建临时文件失败: %v\n", err)
			continue
		}
		
		for _, line := range lines {
			fmt.Fprintln(file, line)
		}
		file.Close()
		fmt.Printf("已保存 %d 行到 %s\n", len(lines), tempFile)
	}
	
	// 合并所有AID文件（包括现有的aid.txt）
	inputFiles := []string{aidFilePath}
	inputFiles = append(inputFiles, tempFiles...)
	
	// 合并到新文件
	newAidFile := aidFilePath + ".new"
	err = mergeAidFiles(newAidFile, inputFiles)
	if err != nil {
		return fmt.Errorf("合并AID文件失败: %v", err)
	}
	
	// 替换原文件
	err = os.Rename(newAidFile, aidFilePath)
	if err != nil {
		return fmt.Errorf("替换aid.txt失败: %v", err)
	}
	
	// 清理临时文件
	for _, tempFile := range tempFiles {
		os.Remove(tempFile)
	}
	
	fmt.Printf("AID文件更新完成！\n")
	return nil
}

// 如果作为独立程序运行
func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := updateAidFile(); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("用法: go run fetch_aids.go update")
		fmt.Println("或者: ./fetch_aids update")
		fmt.Println("")
		fmt.Println("功能:")
		fmt.Println("  更新aid.txt文件，从多个源获取和合并AID")
		fmt.Println("")
		fmt.Println("注意:")
		fmt.Println("  1. 程序会自动备份现有的aid.txt文件")
		fmt.Println("  2. 合并时会自动去重，保留更详细的描述")
		fmt.Println("  3. 需要在fetch_aids.go中配置AID数据源URL")
	}
}

