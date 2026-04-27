package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

func getCopyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <notebook> <doc-path>",
		Short: "复制文档",
		Long: `复制文档（在同一目录下创建副本）。

示例：
  siyuan-cli document copy java "笔记/模板文档"`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := document.CopyDocument(document.CopyOptions{
				NotebookIdentifier: args[0],
				Path:               args[1],
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	return cmd
}
