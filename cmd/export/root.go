package export

import (
	"fmt"

	exportlogic "github.com/cicbyte/siyuan-cli/internal/logic/export"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

var exportDocFormat string
var exportDocOutput string
var exportNbFormat string
var exportNbOutput string

func GetExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "导出文档或笔记本",
		Long: `导出思源笔记文档或笔记本为多种格式。

支持格式：
- md: Markdown 文本
- html: HTML 网页
- docx: Word 文档
- sy: 思源格式`,
	}
	cmd.AddCommand(getDocCommand())
	cmd.AddCommand(getNotebookCommand())
	return cmd
}

func getDocCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc <doc-id | notebook> [doc-path]",
		Short: "导出文档",
		Long: `导出单个文档。支持直接传文档 ID 或 笔记本/文档路径。

示例：
  siyuan-cli export doc 20240101120000-xxx --format html -o output.html
  siyuan-cli export doc java "笔记/重要" --format md -o 重要.md`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var opts exportlogic.ExportDocOptions
			opts.Format = exportDocFormat
			opts.OutputFile = exportDocOutput
			if len(args) == 1 && document.IsDocID(args[0]) {
				opts.DocID = args[0]
			} else if len(args) == 2 {
				opts.NotebookIdentifier = args[0]
				opts.Path = args[1]
			} else {
				fmt.Println("❌ 错误: 请提供文档 ID 或 笔记本和文档路径")
				fmt.Println("💡 使用方法: siyuan-cli export doc <doc-id> --format md -o output.md")
				fmt.Println("   或: siyuan-cli export doc <笔记本> <文档路径> --format md -o output.md")
				return
			}
			if err := exportlogic.ExportDoc(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&exportDocFormat, "format", "f", "md", "导出格式 (md|html|docx)")
	cmd.Flags().StringVarP(&exportDocOutput, "output", "o", "", "输出文件路径")
	return cmd
}

func getNotebookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notebook <notebook>",
		Short: "导出笔记本",
		Long: `导出整个笔记本为 zip 压缩包。

示例：
  siyuan-cli export notebook java --format md -o ./backup/
  siyuan-cli export notebook java --format sy -o ./backup/`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := exportlogic.ExportNotebook(exportlogic.ExportNotebookOptions{
				NotebookIdentifier: args[0],
				Format:             exportNbFormat,
				OutputDir:          exportNbOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.MarkFlagRequired("output")
	cmd.Flags().StringVarP(&exportNbFormat, "format", "f", "md", "导出格式 (md|sy)")
	cmd.Flags().StringVarP(&exportNbOutput, "output", "o", "", "输出目录")
	return cmd
}
