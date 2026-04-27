package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

var (
	createMdPath    string
	createMdTitle   string
	createMdContent string
	createMdFile    string
)

func getCreateMdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createMd [notebook-name-or-id]",
		Short: "使用Markdown创建文档",
		Long: `使用Markdown内容创建新的思源笔记文档。

参数：
  notebook-name-or-id: 目标笔记本名称或ID（必需）

选项：
  --path, -p: 文档路径，不包含文件名（如 "/foo/bar"），未提供时进入TUI模式选择
  --title, -t: 文档标题，默认使用Markdown内容的第一个标题
  --content, -c: Markdown文本内容
  --file, -F: 包含Markdown内容的文件路径
  --format: 输出格式 (table|json)，全局默认为 table

智能匹配规则：
  如果输入符合ID格式（如 20231201120000-abc123），直接使用ID
  否则进行名称匹配：精确匹配 → 不区分大小写匹配 → 包含匹配
  如果找到多个匹配项，会列出所有匹配的笔记本供参考

输入方式（三种方式只能使用一种）：
  --content 参数直接提供Markdown文本
  --file 参数从文件读取Markdown内容
  管道输入（如 cat file.md | siyuan-cli document createMd notebook）

示例：
  siyuan-cli document createMd python --content "# 标题\n\n这是内容"
  siyuan-cli document createMd "工作笔记" --file readme.md --path "/项目文档"
  cat readme.md | siyuan-cli document createMd "工作笔记" --title "API文档"
  siyuan-cli document createMd python --title "API文档" --file api.md --format json`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID")
				fmt.Println("💡 使用方法: siyuan-cli document createMd <笔记本名称或ID> [选项]")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				return
			}

			opts := document.CreateMdOptions{
				NotebookIdentifier: args[0],
				Path:               createMdPath,
				Title:              createMdTitle,
				Markdown:           createMdContent,
				MarkdownFile:       createMdFile,
			}

			if err := document.CreateMdWithMarkdown(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&createMdPath, "path", "p", "", "文档路径，不包含文件名，未提供时进入TUI模式选择")
	cmd.Flags().StringVarP(&createMdTitle, "title", "t", "", "文档标题")
	cmd.Flags().StringVarP(&createMdContent, "content", "c", "", "Markdown文本内容")
	cmd.Flags().StringVarP(&createMdFile, "file", "F", "", "包含Markdown内容的文件路径")

	cmd.RegisterFlagCompletionFunc("file", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"md", "markdown"}, cobra.ShellCompDirectiveFilterFileExt
	})

	return cmd
}
