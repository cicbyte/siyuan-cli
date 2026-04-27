package synccmd

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/synccmd"
	"github.com/spf13/cobra"
)

func GetSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步管理",
		Long: `思源笔记同步工具。

可以用来：
- 查看同步状态
- 立即执行同步`,
	}
	cmd.AddCommand(getStatusCommand())
	cmd.AddCommand(getNowCommand())
	return cmd
}

func getStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看同步状态",
		Long:  "查看思源笔记的当前同步状态。",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := synccmd.SyncStatus(); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	return cmd
}

func getNowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "now",
		Short: "立即同步",
		Long:  "立即执行思源笔记同步。",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := synccmd.SyncNow(); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	return cmd
}
