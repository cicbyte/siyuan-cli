package auth

import "github.com/spf13/cobra"

func GetAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "思源笔记连接管理",
		Long: `管理 siyuan-cli 与思源笔记的连接认证。

示例:
  siyuan-cli auth login   # 交互式配置思源笔记连接
  siyuan-cli auth status  # 查看当前连接状态
  siyuan-cli auth logout  # 断开连接并清除凭据`,
	}
	cmd.AddCommand(getLoginCommand())
	cmd.AddCommand(getStatusCommand())
	cmd.AddCommand(getLogoutCommand())
	return cmd
}
