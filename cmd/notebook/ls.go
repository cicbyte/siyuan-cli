package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

var (
	showClosed bool
	sortBy     string
	outputFile string
)

func getListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出笔记本",
		Long: `列出所有思源笔记笔记本。

支持多种输出格式：
- 表格输出（默认）：简洁的终端表格
- JSON输出：使用 --format json 参数

可以使用过滤选项：
- --closed: 显示已关闭的笔记本
- --sort: 排序方式（name, id, created, updated）
- --format, -f: 输出格式（table, json）
- --o, --output: 导出到指定文件`,
		Run: func(cmd *cobra.Command, args []string) {
			opts := notebook.Options{
				ShowClosed: showClosed,
				SortBy:     sortBy,
				OutputFile: outputFile,
			}

			if err := notebook.ListNotebooks(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().BoolVar(&showClosed, "closed", false, "显示已关闭的笔记本")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "排序方式 (name, id, created, updated)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "导出到指定文件")

	cmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"name", "id", "created", "updated"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
