package block

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/block"
	"github.com/spf13/cobra"
)

var blockGetOutput string

func getGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <block-id>",
		Short: "查看块信息",
		Long:  "查看块的详细信息。",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := block.GetBlock(block.GetBlockOptions{
				ID:         args[0],
				OutputFile: blockGetOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&blockGetOutput, "output", "o", "", "导出到文件")
	return cmd
}

var blockSourceOutput string

func getSourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source <block-id>",
		Short: "获取块源码",
		Long:  "获取块的 kramdown 源码。",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := block.GetBlockSource(block.SourceOptions{
				ID:         args[0],
				OutputFile: blockSourceOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&blockSourceOutput, "output", "o", "", "导出到文件")
	return cmd
}
