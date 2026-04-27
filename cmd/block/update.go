package block

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/block"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/spf13/cobra"
)

var blockUpdateContent string
var blockAppendContent string
var blockAppendType string
var blockDeleteForce bool

func getUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <block-id>",
		Short: "更新块内容",
		Long: `更新块内容。支持通过 --content 参数或管道传入内容。

示例：
  siyuan-cli block update 20240101120000-xxx --content "新内容"
  echo "新内容" | siyuan-cli block update 20240101120000-xxx`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			content := blockUpdateContent
			if content == "" {
				content, _ = output.ReadPipeOrFile("")
			}
			if err := block.UpdateBlock(block.UpdateOptions{
				ID:      args[0],
				Content: content,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&blockUpdateContent, "content", "c", "", "新内容")
	return cmd
}

func getAppendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "append <doc-id | notebook/doc-path>",
		Short: "追加内容到文档",
		Long: `追加内容到文档末尾。支持通过 --content 参数或管道传入内容。

示例：
  siyuan-cli block append 20240101120000-xxx --content "## 新增段落"
  siyuan-cli block append java/"日记/2024-01-01" --content "## 新增段落"
  echo "**加粗文本**" | siyuan-cli block append 20240101120000-xxx`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			content := blockAppendContent
			if content == "" {
				content, _ = output.ReadPipeOrFile("")
			}
			// 如果参数不包含 /，视为文档 ID
			arg := args[0]
			var opts block.AppendOptions
			if document.IsDocID(arg) {
				opts.DocID = arg
			} else {
				nb, path := parseNotebookPath(arg)
				opts.NotebookIdentifier = nb
				opts.Path = path
			}
			opts.Content = content
			opts.DataType = blockAppendType
			if err := block.AppendBlock(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&blockAppendContent, "content", "c", "", "内容")
	cmd.Flags().StringVar(&blockAppendType, "type", "markdown", "内容类型 (markdown|dom)")
	return cmd
}

func getDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <block-id>",
		Short: "删除块",
		Long:  "删除指定的块。",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := block.DeleteBlock(block.DeleteOptions{
				ID:    args[0],
				Force: blockDeleteForce,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().BoolVar(&blockDeleteForce, "force", false, "强制删除")
	return cmd
}

func parseNotebookPath(s string) (string, string) {
	parts := splitN(s, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

func splitN(s string, sep string, n int) []string {
	var result []string
	start := 0
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep, start)
		if idx == -1 {
			break
		}
		result = append(result, s[start:idx])
		start = idx + len(sep)
	}
	result = append(result, s[start:])
	return result
}

func indexOf(s, substr string, fromIndex int) int {
	for i := fromIndex; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
