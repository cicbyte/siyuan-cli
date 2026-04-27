package auth

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/utils"
	"github.com/spf13/cobra"
)

func getLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "断开思源笔记连接",
		Long: `清除本地保存的思源笔记连接凭据。

此操作仅清除本地配置，不会影响思源笔记服务端。`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := common.GetAppConfig()

			if !cfg.SiYuan.Enabled {
				fmt.Println("当前未连接思源笔记")
				return
			}

			newCfg := *cfg
			newCfg.SiYuan.Enabled = false
			newCfg.SiYuan.ApiToken = ""

			utils.ConfigInstance.SaveConfig(&newCfg)
			common.SetAppConfig(&newCfg)

			fmt.Println("已断开思源笔记连接")
		},
	}
}
