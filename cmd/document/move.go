package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

func getMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <notebook> <src-path> <dest-path>",
		Short: "移动文档",
		Long: `移动文档到新位置。

支持跨笔记本移动，使用 notebook2:/路径 格式指定目标。

示例：
  siyuan-cli document move java "旧目录/笔记" "新目录/笔记"
  siyuan-cli document move java "笔记" blog:/接收目录`,
		Args: cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			if err := document.MoveDocument(document.MoveOptions{
				NotebookIdentifier: args[0],
				SrcPath:            args[1],
				DestPath:           args[2],
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	return cmd
}
