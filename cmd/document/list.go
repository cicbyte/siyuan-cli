package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

var (
	docPath   string
	docSort   int
	docDepth  int
	docOutput string
)

func getListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [notebook-name-or-id]",
		Short: "列出文档树",
		Long: `列出指定思源笔记笔记本的文档树结构。

	参数：
			- notebook-name-or-id: 要列出的笔记本名称或ID（必需）

	选项：
			- --path, -p: 文档路径，从根路径开始（如 "/a/b"）。如果为空，则返回整个笔记本的文档树
			- --sort, -s: 排序方式（0：按名称，1：按更新时间，2：按创建时间，3：自定义）
			- --depth, -d: 显示深度（默认1，0表示全部展开）
			- --format, -f: 输出格式 (table|json)，默认为 table
			- --output, -o: 导出到指定文件（自动使用JSON格式）

	智能匹配规则：
			- 如果输入符合ID格式（如 20231201120000-abc123），直接使用ID
			- 否则进行名称匹配：精确匹配 → 不区分大小写匹配 → 包含匹配
			- 如果找到多个匹配项，会列出所有匹配的笔记本供参考

	输出格式：
			- table（默认）: 树形展示文档目录结构
			- json: JSON格式，便于程序处理和集成
			- 文件导出: 当使用 --output 参数时，自动强制使用JSON格式

	示例：
			- siyuan-cli document list python                    # 根目录（默认depth=1）
			- siyuan-cli document list python --depth 2          # 展开到第2层
			- siyuan-cli document list python --depth 0          # 全部展开
			- siyuan-cli document list "工作笔记" --format json  # JSON格式输出
			- siyuan-cli document list "python" --path "/基础"   # 列出指定路径下的文档
			- siyuan-cli notebook list  # 查看所有笔记本`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID")
				fmt.Println("💡 使用方法: siyuan-cli document list <笔记本名称或ID>")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				return
			}

			opts := document.ListOptions{
				NotebookIdentifier: args[0],
				Path:               docPath,
				Sort:               docSort,
				Depth:              docDepth,
				OutputFile:         docOutput,
			}

			if err := document.ListDocuments(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&docPath, "path", "p", "", "文档路径，从根路径开始")
	cmd.Flags().IntVarP(&docSort, "sort", "s", 0, "排序方式 (0:名称, 1:更新时间, 2:创建时间, 3:自定义)")
	cmd.Flags().IntVarP(&docDepth, "depth", "d", 1, "显示深度 (1:仅根目录, 0:全部展开)")
	cmd.Flags().StringVarP(&docOutput, "output", "o", "", "导出到指定文件（自动使用JSON格式）")

	cmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"0", "1", "2", "3"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
