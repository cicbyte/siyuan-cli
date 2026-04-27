package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/models"
	"go.yaml.in/yaml/v3"
)

var ConfigInstance = Config{}

type Config struct {
	HomeDir      string
	AppSeriesDir string
	AppDir       string
	ConfigDir    string
	ConfigPath   string
	LogDir       string
	LogPath      string
}

func (c *Config) GetHomeDir() string {
	if c.HomeDir != "" {
		return c.HomeDir
	}
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}
	c.HomeDir = usr.HomeDir
	return c.HomeDir
}

func (c *Config) GetAppSeriesDir() string {
	if c.AppSeriesDir != "" {
		return c.AppSeriesDir
	}
	c.AppSeriesDir = c.GetHomeDir() + "/.cicbyte"
	return c.AppSeriesDir
}

func (c *Config) GetAppDir() string {
	if c.AppDir != "" {
		return c.AppDir
	}
	c.AppDir = c.GetAppSeriesDir() + "/siyuan-cli"
	return c.AppDir
}

func (c *Config) GetConfigDir() string {
	if c.ConfigDir != "" {
		return c.ConfigDir
	}
	c.ConfigDir = c.GetAppDir() + "/config"
	return c.ConfigDir
}
func (c *Config) GetConfigPath() string {
	if c.ConfigPath != "" {
		return c.ConfigPath
	}
	c.ConfigPath = c.GetConfigDir() + "/config.yaml"
	return c.ConfigPath
}

func (c *Config) GetLogDir() string {
	if c.LogDir == "" {
		c.LogDir = filepath.Join(c.GetAppDir(), "logs")
	}
	return c.LogDir
}

func (c *Config) GetLogPath() string {
	if c.LogPath == "" {
		now := time.Now().Format("20060102")
		c.LogPath = filepath.Join(c.GetLogDir(), fmt.Sprintf("siyuan-cli_log_%s.log", now))
	}
	return c.LogPath
}

func (c *Config) LoadConfig() *models.AppConfig {
	config_path := c.GetConfigPath()

	if _, err := os.Stat(config_path); os.IsNotExist(err) {
		defaultConfig := GetDefaultConfig()
		data, err := yaml.Marshal(defaultConfig)
		if err == nil {
			_ = os.WriteFile(config_path, data, 0644)
		}
		return defaultConfig
	}

	data, err := os.ReadFile(config_path)
	if err != nil {
		return GetDefaultConfig()
	}

	var config models.AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return GetDefaultConfig()
	}

	return &config
}

func (c *Config) SaveConfig(config *models.AppConfig) {
	config_path := c.GetConfigPath()
	data, err := yaml.Marshal(config)
	if err != nil {
		return
	}
	os.WriteFile(config_path, data, 0644)
}

func GetDefaultConfig() *models.AppConfig {
	config := &models.AppConfig{}

	config.AI.Provider = "ollama"
	config.AI.BaseURL = "http://127.0.0.1:11434"
	config.AI.Model = "gemma4:e4b"
	config.AI.MaxTokens = 2048
	config.AI.Temperature = 0.8
	config.AI.Timeout = 120

	config.Log.Level = "info"
	config.Log.MaxSize = 10
	config.Log.MaxBackups = 30
	config.Log.MaxAge = 30
	config.Log.Compress = true

	config.SiYuan.BaseURL = "http://127.0.0.1:6806"
	config.SiYuan.ApiToken = ""
	config.SiYuan.Timeout = 30
	config.SiYuan.UserAgent = "siyuan-cli/1.0"
	config.SiYuan.RetryCount = 3
	config.SiYuan.Enabled = false
	config.Output.Format = "table"

	return config
}

func (c *Config) GetSiYuanConfig() *models.AppConfig {
	return common.GetAppConfig()
}

func (c *Config) IsSiYuanEnabled() bool {
	config := c.GetSiYuanConfig()
	return config != nil && config.SiYuan.Enabled && config.SiYuan.BaseURL != ""
}

func (c *Config) ValidateSiYuanConfig() error {
	config := c.GetSiYuanConfig()
	if config == nil {
		return fmt.Errorf("配置未加载")
	}
	if config.SiYuan.BaseURL == "" {
		return fmt.Errorf("思源笔记基础URL不能为空")
	}
	return nil
}
