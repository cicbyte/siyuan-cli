package notebook

import "github.com/spf13/cobra"

func GetNotebookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notebook",
		Short: "笔记本管理",
		Long: `思源笔记笔记本管理工具。

可以用来：
- 列出所有笔记本
- 查看笔记本详情
- 创建、删除笔记本
- 管理笔记本状态`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getOpenCommand())
	cmd.AddCommand(getCloseCommand())
	cmd.AddCommand(getCreateCommand())
	cmd.AddCommand(getRenameCommand())
	cmd.AddCommand(getDeleteCommand())
	cmd.AddCommand(getGetConfCommand())
	cmd.AddCommand(getSetConfCommand())
	return cmd
}
