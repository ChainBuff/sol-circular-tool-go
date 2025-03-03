package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置结构
type Config struct {
	JUPITER_API_URLS []string `json:"JUPITER_API_URLS"`
	CIRCULAR_API_URL string   `json:"CIRCULAR_API_URL"`
	APIKey          string   `json:"APIKey"`
}

// Load 从config.json加载配置
func Load() (*Config, error) {
	// 尝试多个可能的配置文件位置
	configPaths := []string{
		"config.json",                     // 当前目录
		"../config.json",                  // 上级目录
		filepath.Join(os.Getenv("HOME"), ".sol-circular-tool", "config.json"), // 用户目录
	}
	
	var configData []byte
	var readErr error
	
	for _, path := range configPaths {
		configData, readErr = os.ReadFile(path)
		if readErr == nil {
			fmt.Printf("使用配置文件: %s\n", path)
			break
		}
	}
	
	if readErr != nil {
		return nil, fmt.Errorf("无法找到或读取配置文件: %w", readErr)
	}
	
	// 解析JSON
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
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