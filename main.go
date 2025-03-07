package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	
	"sol-circular-tool/config"
	"sol-circular-tool/services"
	"sol-circular-tool/models"
)

// 显示帮助信息
func showHelp() {
	fmt.Println("用法: sol-circular-tool -j <Jupiter API URL> -k <API Key> [选项]")
	fmt.Println("选项:")
	fmt.Println("  -h, --help                        显示帮助信息")
	fmt.Println("  -j, --jupiter <URL>               指定Jupiter API URL (必填)")
	fmt.Println("  -k, --key <KEY>                   指定API密钥 (必填)")
	fmt.Println("  -i, --dex-program-ids <ID1,ID2,...>   只包含指定的DEX程序ID (与exclude-dex-program-ids互斥)")
	fmt.Println("  -e, --exclude-dex-program-ids <ID1,ID2,...>  排除指定的DEX程序ID (与dex-program-ids互斥)")
	os.Exit(0)
}

func main() {
	// 定义命令行参数
	var showHelpFlag bool
	var jupiterURL string
	var apiKey string
	var dexProgramIDsStr string
	var excludeDexProgramIDsStr string
	
	flag.BoolVar(&showHelpFlag, "h", false, "显示帮助信息")
	flag.StringVar(&jupiterURL, "j", "", "指定Jupiter API URL (必填)")
	flag.StringVar(&apiKey, "k", "", "指定API密钥 (必填)")
	flag.StringVar(&dexProgramIDsStr, "i", "", "只包含指定的DEX程序ID，逗号分隔")
	flag.StringVar(&excludeDexProgramIDsStr, "e", "", "排除指定的DEX程序ID，逗号分隔")
	
	// 添加长选项别名
	flag.BoolVar(&showHelpFlag, "help", false, "显示帮助信息")
	flag.StringVar(&jupiterURL, "jupiter", "", "指定Jupiter API URL (必填)")
	flag.StringVar(&apiKey, "key", "", "指定API密钥 (必填)")
	flag.StringVar(&dexProgramIDsStr, "dex-program-ids", "", "只包含指定的DEX程序ID，逗号分隔")
	flag.StringVar(&excludeDexProgramIDsStr, "exclude-dex-program-ids", "", "排除指定的DEX程序ID，逗号分隔")
	
	// 解析命令行参数
	flag.Parse()
	
	// 如果指定了帮助标志或参数不足，显示帮助信息并退出
	if showHelpFlag || jupiterURL == "" || apiKey == "" {
		showHelp()
	}
	
	// 检查互斥参数
	if dexProgramIDsStr != "" && excludeDexProgramIDsStr != "" {
		fmt.Println("错误: --dex-program-ids 和 --exclude-dex-program-ids 参数不能同时使用")
		os.Exit(1)
	}
	
	// 解析DEX程序ID
	var dexProgramIDs []string
	var excludeDexProgramIDs []string
	
	if dexProgramIDsStr != "" {
		dexProgramIDs = strings.Split(dexProgramIDsStr, ",")
		for i, id := range dexProgramIDs {
			dexProgramIDs[i] = strings.TrimSpace(id)
		}
		fmt.Printf("只包含以下DEX程序ID: %s\n", strings.Join(dexProgramIDs, ", "))
	}
	
	if excludeDexProgramIDsStr != "" {
		excludeDexProgramIDs = strings.Split(excludeDexProgramIDsStr, ",")
		for i, id := range excludeDexProgramIDs {
			excludeDexProgramIDs[i] = strings.TrimSpace(id)
		}
		fmt.Printf("排除以下DEX程序ID: %s\n", strings.Join(excludeDexProgramIDs, ", "))
	}
	
	// 创建配置
	cfg := config.NewConfig(jupiterURL, apiKey, dexProgramIDs, excludeDexProgramIDs)
	
	// 创建结果通道
	resultChan := make(chan *struct {
		URL string
		Data []models.InputMarketData
		Error error
	}, 1)
	
	// 启动Jupiter API处理
	var wg sync.WaitGroup
	wg.Add(1)
	go services.ProcessJupiterAPI(cfg.JUPITER_API_URL, "https://pro.circular.bot/market/cache", cfg.APIKey, &wg, resultChan)
	
	// 等待处理完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 收集结果
	var successData []models.InputMarketData
	var successURL string
	for result := range resultChan {
		if result.Error == nil {
			successData = result.Data
			successURL = result.URL
			break
		}
	}
	
	// 检查是否有成功的结果
	if successData == nil {
		log.Fatal("无法从Jupiter API获取数据")
	}
	
	// 过滤数据
	if len(cfg.DexProgramIDs) > 0 || len(cfg.ExcludeDexProgramIDs) > 0 {
		successData = filterMarketData(successData, cfg.DexProgramIDs, cfg.ExcludeDexProgramIDs)
		fmt.Printf("过滤后剩余 %d 条市场数据\n", len(successData))
	}
	
	// 处理数据
	fmt.Println("正在处理数据...")
	outputData := services.ProcessMarketData(successData)
	
	// 添加重试逻辑
	maxRetries := 3
	for retries := 0; retries < maxRetries; retries++ {
		err := services.SubmitMarketData(outputData, successURL)
		if err == nil {
			break
		}
		
		fmt.Printf("提交数据失败(尝试 %d/%d): %v\n", retries+1, maxRetries, err)
		if retries < maxRetries-1 {
			time.Sleep(5 * time.Second) // 延迟后重试
		}
	}

	// 记录完成情况
	fmt.Printf("处理完成! 总共处理了 %d 条数据记录\n", len(outputData))
}

// filterMarketData 根据DEX程序ID过滤市场数据
func filterMarketData(data []models.InputMarketData, includeIDs, excludeIDs []string) []models.InputMarketData {
	if len(includeIDs) == 0 && len(excludeIDs) == 0 {
		return data
	}
	
	var result []models.InputMarketData
	
	// 创建包含和排除的映射，以便快速查找
	includeMap := make(map[string]bool)
	excludeMap := make(map[string]bool)
	
	for _, id := range includeIDs {
		includeMap[id] = true
	}
	
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}
	
	// 过滤数据
	for _, item := range data {
		// 如果有包含列表，检查owner是否在列表中
		if len(includeMap) > 0 {
			if includeMap[item.Owner] {
				result = append(result, item)
			}
		} else if len(excludeMap) > 0 {
			// 如果有排除列表，检查owner是否不在列表中
			if !excludeMap[item.Owner] {
				result = append(result, item)
			}
		}
	}
	
	return result
} 