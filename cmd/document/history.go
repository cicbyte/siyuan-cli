package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

var historyQuery string
var historyOutput string
var rollbackTo string

func getHistoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [notebook] [doc-path]",
		Short: "查看文档历史",
		Long: `查看文档历史版本列表。

示例：
  siyuan-cli document history java "笔记/重要文档"
  siyuan-cli document history --query RAG
  siyuan-cli document history java "笔记" -o history.json`,
		Args: cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var notebook, path string
			if len(args) >= 1 {
				notebook = args[0]
			}
			if len(args) >= 2 {
				path = args[1]
			}
			if err := document.GetDocumentHistory(document.HistoryOptions{
				NotebookIdentifier: notebook,
				Path:               path,
				Query:              historyQuery,
				OutputFile:         historyOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&historyQuery, "query", "q", "", "搜索关键词")
	cmd.Flags().StringVarP(&historyOutput, "output", "o", "", "导出到文件")
	return cmd
}

func getRollbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback --notebook <笔记本> --to <历史路径>",
		Short: "回滚文档",
		Long: `回滚文档到指定历史版本。

示例：
  siyuan-cli document history --query RAG
  siyuan-cli document rollback --notebook java --to 20240101120000-xxx`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			notebook, _ := cmd.Flags().GetString("notebook")
			if rollbackTo == "" {
				fmt.Println("❌ 错误: 请使用 --to 指定回滚目标历史路径")
				fmt.Println("💡 使用 'siyuan-cli document history' 查看可用版本")
				return
			}
			if err := document.RollbackDocument(document.RollbackOptions{
				NotebookIdentifier: notebook,
				HistoryPath:        rollbackTo,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringP("notebook", "n", "", "笔记本名称或 ID")
	cmd.Flags().StringVar(&rollbackTo, "to", "", "回滚目标历史路径（从 history 命令获取的 ID）")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}
