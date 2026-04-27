package common

import (
	"sync/atomic"

	"github.com/cicbyte/siyuan-cli/internal/models"
)

var appConfig atomic.Pointer[models.AppConfig]

func SetAppConfig(cfg *models.AppConfig) {
	appConfig.Store(cfg)
}

func GetAppConfig() *models.AppConfig {
	return appConfig.Load()
}
