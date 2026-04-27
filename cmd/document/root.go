package document

import "github.com/spf13/cobra"

func GetDocumentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "document",
		Short: "文档管理",
		Long: `思源笔记文档管理工具。

可以用来：
- 列出文档树结构
- 查看文档详细信息
- 管理文档和文件夹
- 使用Markdown创建新文档
- 重命名现有文档
- 删除现有文档`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getGetCommand())
	cmd.AddCommand(getCreateMdCommand())
	cmd.AddCommand(getRenameCommand())
	cmd.AddCommand(getDeleteCommand())
	cmd.AddCommand(getMoveCommand())
	cmd.AddCommand(getCopyCommand())
	cmd.AddCommand(getOutlineCommand())
	cmd.AddCommand(getDailyCommand())
	cmd.AddCommand(getHistoryCommand())
	cmd.AddCommand(getRollbackCommand())
	return cmd
}
