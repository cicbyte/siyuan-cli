package search

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/search"
	"github.com/spf13/cobra"
)

var searchDocNotebook string
var searchDocLimit int
var searchDocOutput string

func getDocCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc <keyword>",
		Short: "搜索文档",
		Long: `搜索文档标题和路径。

示例：
  siyuan-cli search doc "AI"
  siyuan-cli search doc "面试" --notebook java --limit 20
  siyuan-cli search doc "面试" -o results.json`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts := search.SearchDocOptions{
				Keyword:    args[0],
				Notebook:   searchDocNotebook,
				Limit:      searchDocLimit,
				OutputFile: searchDocOutput,
			}
			if err := search.SearchDoc(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&searchDocNotebook, "notebook", "n", "", "限定笔记本范围")
	cmd.Flags().IntVarP(&searchDocLimit, "limit", "l", 0, "最大结果数")
	cmd.Flags().StringVarP(&searchDocOutput, "output", "o", "", "导出到文件")
	return cmd
}
