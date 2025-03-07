package config

// Config 配置结构
type Config struct {
	JUPITER_API_URL      string
	APIKey               string
	DexProgramIDs        []string // 包含的DEX程序ID列表
	ExcludeDexProgramIDs []string // 排除的DEX程序ID列表
}

// NewConfig 创建新的配置
func NewConfig(jupiterURL, apiKey string, dexProgramIDs, excludeDexProgramIDs []string) *Config {
	return &Config{
		JUPITER_API_URL:      jupiterURL,
		APIKey:               apiKey,
		DexProgramIDs:        dexProgramIDs,
		ExcludeDexProgramIDs: excludeDexProgramIDs,
	}
} 