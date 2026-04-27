package tag

import "github.com/spf13/cobra"

func GetTagCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "标签管理",
		Long: `思源笔记标签管理工具。

可以用来：
- 列出所有标签
- 搜索标签
- 为文档添加标签
- 移除文档标签`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getSearchCommand())
	cmd.AddCommand(getAddCommand())
	cmd.AddCommand(getRemoveCommand())
	return cmd
}
