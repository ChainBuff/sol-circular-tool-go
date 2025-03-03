package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"sol-circular-tool/models"
)

var lastCircularAPICall time.Time
var circularAPIMutex sync.Mutex

// checkCircularAPIDelay 检查并等待API调用延迟
func checkCircularAPIDelay() {
	circularAPIMutex.Lock()
	defer circularAPIMutex.Unlock()
	
	// 计算距离上次调用的时间
	elapsed := time.Since(lastCircularAPICall)
	if elapsed < time.Minute {
		// 如果不足1分钟，等待剩余时间
		waitTime := time.Minute - elapsed
		fmt.Printf("等待 %.1f 秒以满足API调用间隔要求...\n", waitTime.Seconds())
		time.Sleep(waitTime)
	}
	
	// 更新最后调用时间
	lastCircularAPICall = time.Now()
}

// ProcessMarketData 处理输入的市场数据并转换格式
func ProcessMarketData(inputData []models.InputMarketData) []models.OutputMarketData {
	result := make([]models.OutputMarketData, 0, len(inputData))
	
	for _, item := range inputData {
		output := models.OutputMarketData{
			Address: item.Pubkey,
			Owner:   item.Owner,
		}
		
		// 如果有参数，进行处理
		if item.Params != nil {
			// 创建新的参数映射
			output.Params = make(map[string]string)
			
			// 处理 addressLookupTableAddress
			if item.Params.AddressLookupTableAddress != "" {
				output.AddressLookupTableAddress = item.Params.AddressLookupTableAddress
			}
			
			// 处理 routingGroup
			if item.Params.RoutingGroup != 0 {
				output.Params["routingGroup"] = fmt.Sprintf("%d", item.Params.RoutingGroup)
			}
			
			// 处理 VaultLpMint
			if item.Params.VaultLpMint != nil {
				output.Params["vaultLpMintA"] = item.Params.VaultLpMint.A
				output.Params["vaultLpMintB"] = item.Params.VaultLpMint.B
			}
			
			// 处理 VaultToken
			if item.Params.VaultToken != nil {
				output.Params["vaultTokenA"] = item.Params.VaultToken.A
				output.Params["vaultTokenB"] = item.Params.VaultToken.B
			}
			
			// 处理 Serum 相关字段
			if item.Params.SerumAsks != "" {
				output.Params["serumAsks"] = item.Params.SerumAsks
			}
			if item.Params.SerumBids != "" {
				output.Params["serumBids"] = item.Params.SerumBids
			}
			if item.Params.SerumCoinVaultAccount != "" {
				output.Params["serumCoinVaultAccount"] = item.Params.SerumCoinVaultAccount
			}
			if item.Params.SerumEventQueue != "" {
				output.Params["serumEventQueue"] = item.Params.SerumEventQueue
			}
			if item.Params.SerumPcVaultAccount != "" {
				output.Params["serumPcVaultAccount"] = item.Params.SerumPcVaultAccount
			}
			if item.Params.SerumVaultSigner != "" {
				output.Params["serumVaultSigner"] = item.Params.SerumVaultSigner
			}
			
			// 如果参数映射为空，删除它
			if len(output.Params) == 0 {
				output.Params = nil
			}
		}
		
		result = append(result, output)
	}
	
	return result
}

// FetchMarketData 从Circular API获取市场数据
func FetchMarketData(jupiterURL string, apiKey string, tokens string) ([]models.InputMarketData, error) {
	// 检查API调用延迟
	checkCircularAPIDelay()
	
	// 构建请求URL
	requestURL := fmt.Sprintf("%s?onlyjup=false&tokens=%s", jupiterURL, tokens)
	
	// 打印请求URL用于调试（可选）
	fmt.Printf("正在请求URL: %s\n", requestURL)
	
	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回了非200状态码: %d, URL: %s, 响应内容: %s", 
			resp.StatusCode, requestURL, string(body))
	}
	
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %w", err)
	}
	
	// 解析JSON
	var data []models.InputMarketData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}
	
	return data, nil
}

// SubmitMarketData 将处理后的数据提交到Jupiter API
func SubmitMarketData(outputData []models.OutputMarketData, jupiterURL string) error {
	fmt.Printf("开始提交数据到Jupiter，共 %d 条记录\n", len(outputData))
	
	// 构建 addmarket 路径
	addMarketURL := fmt.Sprintf("%s/add-market", strings.TrimRight(jupiterURL, "/"))
	
	// 逐个发送数据
	for i, item := range outputData {
		fmt.Printf("\n处理第 %d/%d 条数据:\n", i+1, len(outputData))
		fmt.Printf("地址: %s\n", item.Address)
		fmt.Printf("所有者: %s\n", item.Owner)
		
		// 构建Jupiter格式的请求数据
		type JupiterRequest struct {
			Address                  string            `json:"address"`
			Owner                   string            `json:"owner"`
			Params                  map[string]string `json:"params,omitempty"`
			AddressLookupTableAddress *string           `json:"addressLookupTableAddress"` // 使用指针类型以支持null值
		}
		
		jupiterRequest := JupiterRequest{
			Address: item.Address,
			Owner:   item.Owner,
			AddressLookupTableAddress: nil, // 默认为null
		}
		
		// 如果有 addressLookupTableAddress，则设置它
		if item.AddressLookupTableAddress != "" {
			jupiterRequest.AddressLookupTableAddress = &item.AddressLookupTableAddress
		}
		
		// 如果有 Serum 相关参数，添加到请求中
		if item.Params != nil {
			serumParams := make(map[string]string)
			serumFields := []string{
				"serumAsks",
				"serumBids",
				"serumCoinVaultAccount",
				"serumEventQueue",
				"serumPcVaultAccount",
				"serumVaultSigner",
			}
			
			hasSerumParams := false
			for _, field := range serumFields {
				if value, ok := item.Params[field]; ok {
					serumParams[field] = value
					hasSerumParams = true
				}
			}
			
			if hasSerumParams {
				jupiterRequest.Params = serumParams
			}
		}
		
		// 将数据转换为JSON
		jsonData, err := json.Marshal(jupiterRequest)
		if err != nil {
			return fmt.Errorf("序列化数据失败: %w", err)
		}
		
		fmt.Printf("提交到Jupiter: %s\n请求数据: %s\n", addMarketURL, string(jsonData))
		
		// 创建请求
		req, err := http.NewRequest("POST", addMarketURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}
		
		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		
		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("发送请求失败: %w", err)
		}
		
		// 读取响应内容
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		// 检查响应状态
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Printf("提交失败: HTTP %d - %s\n", resp.StatusCode, string(body))
			return fmt.Errorf("Jupiter API返回了错误状态码: %d, 响应内容: %s", resp.StatusCode, string(body))
		}
		
		fmt.Printf("提交成功: %s\n", string(body))
	}
	
	fmt.Println("\n所有数据提交完成")
	return nil
}

// FetchTokens 从Jupiter API获取代币列表
func FetchTokens(baseURL string) (string, error) {
	tokensURL := fmt.Sprintf("%s/tokens", baseURL)
	
	// 创建请求
	req, err := http.NewRequest("GET", tokensURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建tokens请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取tokens失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tokens API返回了非200状态码: %d, URL: %s, 响应内容: %s",
			resp.StatusCode, tokensURL, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取tokens响应内容失败: %w", err)
	}
	
	// 解析JSON数组
	var tokenAddresses []string
	if err := json.Unmarshal(body, &tokenAddresses); err != nil {
		return "", fmt.Errorf("解析token地址JSON失败: %w", err)
	}
	
	// 将数组连接成逗号分隔的字符串
	tokens := strings.Join(tokenAddresses, ",")
	return tokens, nil
} 