package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/cicbyte/siyuan-cli/internal/utils"
	"github.com/spf13/cobra"
)

func getLoginCommand() *cobra.Command {
	var (
		flagURL   string
		flagToken string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "配置思源笔记连接",
		Long: `交互式配置 siyuan-cli 与思源笔记的连接。

会引导你设置：
1. 思源笔记基础 URL（如 http://127.0.0.1:6806）
2. API Token（在思源笔记 设置→关于→API token 中获取）

配置完成后自动测试连接。

也可以通过参数直接指定：
  siyuan-cli auth login --url http://127.0.0.1:6806 --token your-token`,
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			cfg := common.GetAppConfig()

			// URL：始终提示，显示当前值作为默认
			url := flagURL
			if url == "" {
				currentURL := cfg.SiYuan.BaseURL
				if currentURL == "" {
					currentURL = "http://127.0.0.1:6806"
				}
				fmt.Printf("请输入思源笔记地址 (默认 %s): ", currentURL)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if input == "" {
					url = currentURL
				} else {
					url = input
				}
			}
			url = strings.TrimRight(url, "/")

			// Token：始终提示，显示当前值作为默认
			token := flagToken
			if token == "" {
				if cfg.SiYuan.ApiToken != "" {
					fmt.Print("请输入 API Token (当前已配置，直接回车保留): ")
				} else {
					fmt.Print("请输入 API Token: ")
				}
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if input == "" {
					token = cfg.SiYuan.ApiToken
				} else {
					token = input
				}
			}

			if token == "" {
				fmt.Println("API Token 不能为空")
				fmt.Println("请在思源笔记 设置→关于→API token 中获取")
				os.Exit(1)
			}

			// 保存配置
			newCfg := *cfg
			newCfg.SiYuan.BaseURL = url
			newCfg.SiYuan.ApiToken = token
			newCfg.SiYuan.Enabled = true

			// 测试连接
			fmt.Println("\n正在测试连接...")
			client := siyuan.New(url, token)
			_, err := client.ListNotebooks(context.Background())
			if err != nil {
				fmt.Printf("连接失败: %v\n", err)
				fmt.Println("请确认思源笔记正在运行且地址和 Token 正确")
				os.Exit(1)
			}

			utils.ConfigInstance.SaveConfig(&newCfg)
			common.SetAppConfig(&newCfg)

			fmt.Println("连接成功！思源笔记已配置完毕。")
			fmt.Printf("  地址: %s\n", url)
			fmt.Println("  Token: 已配置")
		},
	}

	cmd.Flags().StringVar(&flagURL, "url", "", "思源笔记基础 URL")
	cmd.Flags().StringVar(&flagToken, "token", "", "API Token")
	return cmd
}
