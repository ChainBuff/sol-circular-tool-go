package config

// Config 配置结构
type Config struct {
	JUPITER_API_URL string
	APIKey          string
}

// NewConfig 创建新的配置
func NewConfig(jupiterURL, apiKey string) *Config {
	return &Config{
		JUPITER_API_URL: jupiterURL,
		APIKey:          apiKey,
	}
} 