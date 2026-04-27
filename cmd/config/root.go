package config

import "github.com/spf13/cobra"

func GetConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理思源笔记连接配置",
		Long: `管理 siyuan-cli 与思源笔记的连接配置。

示例:
  siyuan-cli config list     # 列出所有配置
  siyuan-cli config get siyuan.base_url
  siyuan-cli config set siyuan.enabled true`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getGetCommand())
	cmd.AddCommand(getSetCommand())
	return cmd
}

func getListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有配置项",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 实现配置列表
		},
	}
}

func getGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "获取配置项",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 实现配置获取
		},
	}
}

func getSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "设置配置项",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 实现配置设置
		},
	}
}
