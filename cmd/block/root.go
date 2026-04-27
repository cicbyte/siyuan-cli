package block

import "github.com/spf13/cobra"

func GetBlockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "块操作",
		Long: `思源笔记块操作工具。

可以用来：
- 查看块信息
- 获取块源码
- 更新块内容
- 追加内容到文档
- 删除块`,
	}
	cmd.AddCommand(getGetCommand())
	cmd.AddCommand(getSourceCommand())
	cmd.AddCommand(getUpdateCommand())
	cmd.AddCommand(getAppendCommand())
	cmd.AddCommand(getDeleteCommand())
	return cmd
}
