package search

import "github.com/spf13/cobra"

func GetSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "搜索文档或块",
		Long: `思源笔记搜索工具。

可以用来：
- 搜索文档标题
- 全文搜索块内容`,
	}
	cmd.AddCommand(getDocCommand())
	cmd.AddCommand(getBlockCommand())
	return cmd
}
