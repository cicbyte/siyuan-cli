package models

type AppConfig struct {
	Version string `yaml:"version"` // 版本号，用于升级时判断
	AI      struct {
		Provider    string  `yaml:"provider"` // openai/ollama
		BaseURL     string  `yaml:"base_url"`
		ApiKey      string  `yaml:"api_key"`
		Model       string  `yaml:"model"`
		MaxTokens   int     `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     int     `yaml:"timeout"`
	} `yaml:"ai"`

	Log struct {
		Level      string `yaml:"level"`
		MaxSize    int    `yaml:"maxSize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`

	SiYuan struct {
		BaseURL    string `yaml:"base_url"`    // 思源笔记基础 URL，如 "http://127.0.0.1:6806"
		ApiToken   string `yaml:"api_token"`   // 访问令牌（可选）
		Timeout    int    `yaml:"timeout"`     // 请求超时时间（秒）
		UserAgent  string `yaml:"user_agent"`  // 用户代理
		RetryCount int    `yaml:"retry_count"` // 重试次数
		Enabled    bool   `yaml:"enabled"`     // 是否启用思源笔记功能
	} `yaml:"siyuan"`

	Output struct {
		Format string `yaml:"format"` // 输出格式: table / json
	} `yaml:"output"`
}
