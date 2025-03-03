package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	
	"sol-circular-tool/config"
	"sol-circular-tool/services"
	"sol-circular-tool/models"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	
	var wg sync.WaitGroup
	resultChan := make(chan *struct {
		URL string
		Data []models.InputMarketData
		Error error
	}, len(cfg.JUPITER_API_URLS))
	
	// 启动所有Jupiter API的处理线程
	for _, jupiterURL := range cfg.JUPITER_API_URLS {
		wg.Add(1)
		go services.ProcessJupiterAPI(jupiterURL, cfg.CIRCULAR_API_URL, cfg.APIKey, &wg, resultChan)
	}
	
	// 等待所有处理完成
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
			break // 使用第一个成功的结果
		}
	}
	
	// 检查是否有成功的结果
	if successData == nil {
		log.Fatal("所有 Jupiter APIs 都无法获取数据")
	}
	
	// 处理数据
	fmt.Println("正在处理数据...")
	outputData := services.ProcessMarketData(successData)
	
	// 添加重试逻辑
	maxRetries := 3
	for retries := 0; retries < maxRetries; retries++ {
		err = services.SubmitMarketData(outputData, successURL)
		if err == nil {
			break
		}
		
		fmt.Printf("提交数据失败(尝试 %d/%d): %v\n", retries+1, maxRetries, err)
		if retries < maxRetries-1 {
			time.Sleep(5 * time.Second) // 延迟后重试
		}
	}

	if err != nil {
		log.Fatalf("多次尝试提交数据均失败: %v", err)
	}
	
	// 记录完成情况
	fmt.Printf("处理完成! 总共处理了 %d 条数据记录\n", len(outputData))
} 