package siyuan

import (
	"fmt"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/common"
)

// IsSiYuanConfigValid 检查思源笔记配置是否有效
func IsSiYuanConfigValid() bool {
	config := common.GetAppConfig()
	return config != nil && config.SiYuan.Enabled && config.SiYuan.BaseURL != ""
}

// GetDefaultSiYuanClient 获取默认思源笔记客户端（使用全局配置）
func GetDefaultSiYuanClient() (*Client, error) {
	config := common.GetAppConfig()
	if config == nil {
		return nil, fmt.Errorf("应用配置未加载")
	}

	if !config.SiYuan.Enabled {
		return nil, fmt.Errorf("思源笔记功能未启用，请检查配置")
	}

	siyuanConfig := Config{
		BaseURL: config.SiYuan.BaseURL,
		Token:   config.SiYuan.ApiToken,
		Timeout: time.Duration(config.SiYuan.Timeout) * time.Second,
	}

	return NewWithConfig(siyuanConfig), nil
}