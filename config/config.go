package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 配置结构
type Config struct {
	JUPITER_API_URLS []string `json:"JUPITER_API_URLS"`
	CIRCULAR_API_URL string   `json:"CIRCULAR_API_URL"`
	APIKey          string   `json:"APIKey"`
}

// Load 从config.json加载配置
func Load() (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	// 解析JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	// 验证必要的配置项
	if len(config.JUPITER_API_URLS) == 0 {
		return nil, fmt.Errorf("缺少 JUPITER_API_URLS 配置")
	}
	if config.CIRCULAR_API_URL == "" {
		return nil, fmt.Errorf("缺少 CIRCULAR_API_URL 配置")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("缺少 APIKey 配置")
	}
	
	return &config, nil
} 