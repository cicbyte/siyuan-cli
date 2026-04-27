package search

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/search"
	"github.com/spf13/cobra"
)

var searchBlockNotebook string
var searchBlockLimit int
var searchBlockOutput string

func getBlockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block <keyword>",
		Short: "全文搜索块",
		Long: `全文搜索块内容。

示例：
  siyuan-cli search block "goroutine"
  siyuan-cli search block "并发安全" --notebook java --limit 50
  siyuan-cli search block "TODO" -o results.json`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts := search.SearchBlockOptions{
				Keyword:    args[0],
				Notebook:   searchBlockNotebook,
				Limit:      searchBlockLimit,
				OutputFile: searchBlockOutput,
			}
			if err := search.SearchBlock(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&searchBlockNotebook, "notebook", "n", "", "限定笔记本范围")
	cmd.Flags().IntVarP(&searchBlockLimit, "limit", "l", 0, "最大结果数")
	cmd.Flags().StringVarP(&searchBlockOutput, "output", "o", "", "导出到文件")
	return cmd
}
