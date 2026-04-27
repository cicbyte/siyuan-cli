package tag

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/logic/tag"
	"github.com/spf13/cobra"
)

var tagAddTags []string

func getAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <doc-id | notebook> [doc-path]",
		Short: "添加标签",
		Long: `为文档添加标签。支持直接传文档 ID 或 笔记本/文档路径。

示例：
  siyuan-cli tag add 20240101120000-xxx --tag "重要" --tag "TODO"
  siyuan-cli tag add java "笔记/重要" --tag "重要"`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var opts tag.AddTagOptions
			opts.Tags = tagAddTags
			if len(args) == 1 && document.IsDocID(args[0]) {
				opts.DocID = args[0]
			} else if len(args) == 2 {
				opts.NotebookIdentifier = args[0]
				opts.Path = args[1]
			} else {
				fmt.Println("❌ 错误: 请提供文档 ID 或 笔记本和文档路径")
				fmt.Println("💡 使用方法: siyuan-cli tag add <doc-id> --tag \"标签\"")
				fmt.Println("   或: siyuan-cli tag add <笔记本> <文档路径> --tag \"标签\"")
				return
			}
			if err := tag.AddTags(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringArrayVar(&tagAddTags, "tag", nil, "标签（可多次指定）")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

var tagRemoveName string

func getRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <doc-id | notebook> [doc-path]",
		Short: "移除标签",
		Long: `移除文档的标签。支持直接传文档 ID 或 笔记本/文档路径。

示例：
  siyuan-cli tag remove 20240101120000-xxx --tag "TODO"
  siyuan-cli tag remove java "笔记/重要" --tag "TODO"`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var opts tag.RemoveTagOptions
			opts.Tag = tagRemoveName
			if len(args) == 1 && document.IsDocID(args[0]) {
				opts.DocID = args[0]
			} else if len(args) == 2 {
				opts.NotebookIdentifier = args[0]
				opts.Path = args[1]
			} else {
				fmt.Println("❌ 错误: 请提供文档 ID 或 笔记本和文档路径")
				fmt.Println("💡 使用方法: siyuan-cli tag remove <doc-id> --tag \"标签\"")
				fmt.Println("   或: siyuan-cli tag remove <笔记本> <文档路径> --tag \"标签\"")
				return
			}
			if err := tag.RemoveTag(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVar(&tagRemoveName, "tag", "", "要移除的标签")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}
