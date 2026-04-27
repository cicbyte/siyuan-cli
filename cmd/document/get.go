package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

var docGetOutput string

func getGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <notebook-name-or-id> <doc-path>",
		Short: "查看文档内容",
		Long: `查看指定思源笔记文档的 Markdown 内容。

	参数：
			- notebook-name-or-id: 笔记本名称或ID
			- doc-path: 文档路径（人类可读名称，如 "java面试/AI/index"）

	选项：
			- --output, -o: 导出到文件

	示例：
			- siyuan-cli document get java "java面试/AI/index"
			- siyuan-cli document get java "笔记/待办" -o todo.md
			- siyuan-cli --format json document get java "笔记/待办"`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opts := document.GetOptions{
				NotebookIdentifier: args[0],
				Path:               args[1],
				OutputFile:         docGetOutput,
			}
			if err := document.GetDocument(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&docGetOutput, "output", "o", "", "导出到文件")
	return cmd
}
