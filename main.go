package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	fmt.Println("  -h, --help                显示帮助信息")
	fmt.Println("  -j, --jupiter <URL>       指定Jupiter API URL (必填)")
	fmt.Println("  -k, --key <KEY>           指定API密钥 (必填)")
	os.Exit(0)
}

func main() {
	// 定义命令行参数
	var showHelpFlag bool
	var jupiterURL string
	var apiKey string
	
	flag.BoolVar(&showHelpFlag, "h", false, "显示帮助信息")
	flag.StringVar(&jupiterURL, "j", "", "指定Jupiter API URL (必填)")
	flag.StringVar(&apiKey, "k", "", "指定API密钥 (必填)")
	
	// 添加长选项别名
	flag.BoolVar(&showHelpFlag, "help", false, "显示帮助信息")
	flag.StringVar(&jupiterURL, "jupiter", "", "指定Jupiter API URL (必填)")
	flag.StringVar(&apiKey, "key", "", "指定API密钥 (必填)")
	
	// 解析命令行参数
	flag.Parse()
	
	// 如果指定了帮助标志或参数不足，显示帮助信息并退出
	if showHelpFlag || jupiterURL == "" || apiKey == "" {
		showHelp()
	}
	
	// 创建配置
	cfg := config.NewConfig(jupiterURL, apiKey)
	
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