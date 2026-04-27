package auth

import (
	"context"
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/spf13/cobra"
)

func getStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查看思源笔记连接状态",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := common.GetAppConfig()

			if !cfg.SiYuan.Enabled {
				fmt.Println("状态: 未启用")
				fmt.Println("  运行 siyuan-cli auth login 配置连接")
				return
			}

			fmt.Printf("地址: %s\n", cfg.SiYuan.BaseURL)
			if cfg.SiYuan.ApiToken != "" {
				fmt.Println("Token: 已配置")
			} else {
				fmt.Println("Token: 未配置")
			}

			// 测试连接
			client := siyuan.New(cfg.SiYuan.BaseURL, cfg.SiYuan.ApiToken)
			books, err := client.ListNotebooks(context.Background())
			if err != nil {
				fmt.Printf("\n连接失败: %v\n", err)
				fmt.Println("请确认思源笔记正在运行")
				return
			}

			openCount := 0
			for _, b := range books {
				if !b.Closed {
					openCount++
				}
			}
			fmt.Printf("\n连接正常！共 %d 个笔记本（%d 个已打开）\n", len(books), openCount)
		},
	}
}
