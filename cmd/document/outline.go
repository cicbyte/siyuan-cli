package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

func getOutlineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outline <doc-id | notebook> [doc-path]",
		Short: "查看文档大纲",
		Long: `查看文档大纲（标题层级结构）。支持直接传文档 ID 或 笔记本/文档路径。

示例：
  siyuan-cli document outline 20240101120000-xxx
  siyuan-cli document outline java "java面试/AI/index"`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var opts document.OutlineOptions
			if len(args) == 1 && document.IsDocID(args[0]) {
				opts.DocID = args[0]
			} else if len(args) == 2 {
				opts.NotebookIdentifier = args[0]
				opts.Path = args[1]
			} else {
				fmt.Println("❌ 错误: 请提供文档 ID 或 笔记本和文档路径")
				fmt.Println("💡 使用方法: siyuan-cli document outline <doc-id>")
				fmt.Println("   或: siyuan-cli document outline <笔记本> <文档路径>")
				return
			}
			if err := document.GetDocumentOutline(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	return cmd
}
