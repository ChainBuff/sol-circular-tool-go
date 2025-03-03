package main

import (
	"fmt"
	"log"
	"sync"
	
	"sol-circular-tool/config"
	"sol-circular-tool/services"
	"sol-circular-tool/models"
)

func processJupiterAPI(jupiterURL string, cfg *config.Config, wg *sync.WaitGroup, resultChan chan<- *struct {
	URL string
	Data []models.InputMarketData
	Error error
}) {
	defer wg.Done()
	
	// 从API获取tokens
	fmt.Printf("正在从 %s 获取代币列表...\n", jupiterURL)
	tokens, err := services.FetchTokens(jupiterURL)
	if err != nil {
		fmt.Printf("从 %s 获取代币列表失败: %v\n", jupiterURL, err)
		resultChan <- &struct {
			URL string
			Data []models.InputMarketData
			Error error
		}{URL: jupiterURL, Error: err}
		return
	}
	fmt.Printf("从 %s 成功获取代币列表\n", jupiterURL)
	
	// 获取市场数据
	fmt.Printf("正在从 Circular API 获取市场数据...\n")
	inputData, err := services.FetchMarketData(cfg.CIRCULAR_API_URL, cfg.APIKey, tokens)
	if err != nil {
		fmt.Printf("从 Circular API 获取市场数据失败: %v\n", err)
		resultChan <- &struct {
			URL string
			Data []models.InputMarketData
			Error error
		}{URL: jupiterURL, Error: err}
		return
	}
	
	fmt.Printf("从 %s 成功获取 %d 条市场数据记录\n", jupiterURL, len(inputData))
	resultChan <- &struct {
		URL string
		Data []models.InputMarketData
		Error error
	}{URL: jupiterURL, Data: inputData, Error: nil}
}

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
		go processJupiterAPI(jupiterURL, cfg, &wg, resultChan)
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
	
	// 提交处理后的数据
	fmt.Println("正在提交处理后的数据...")
	err = services.SubmitMarketData(outputData, successURL)
	if err != nil {
		log.Fatalf("提交数据失败: %v", err)
	}
	
	// 记录完成情况
	fmt.Printf("处理完成! 总共处理了 %d 条数据记录\n", len(outputData))
} 