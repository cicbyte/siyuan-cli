package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

var dailyNotebook string

func getDailyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daily",
		Short: "创建今日日记",
		Long: `创建今日日记。

示例：
  siyuan-cli document daily --notebook 日记本
  siyuan-cli document daily`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := document.CreateDailyNote(document.DailyOptions{
				NotebookIdentifier: dailyNotebook,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&dailyNotebook, "notebook", "n", "", "笔记本名称或ID")
	return cmd
}
