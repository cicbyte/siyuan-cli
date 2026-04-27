package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

func getOpenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "open [notebook-name-or-id]",
		Short: "打开笔记本",
		Long: `打开指定的思源笔记笔记本，支持名称或ID。

参数：
- notebook-name-or-id: 要打开的笔记本名称或ID（必需）

智能匹配规则：
- 如果输入符合ID格式（如 20231201120000-abc123），直接使用ID
- 否则进行名称匹配：精确匹配 → 不区分大小写匹配 → 包含匹配
- 如果找到多个匹配项，会列出所有匹配的笔记本供参考

示例：
- siyuan-cli notebook open ai                    # 使用名称
- siyuan-cli notebook open 20251105164527-ezlspgg  # 使用ID
- siyuan-cli notebook list  # 查看所有笔记本`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID")
				fmt.Println("💡 使用方法: siyuan-cli notebook open <笔记本名称或ID>")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				return
			}

			opts := notebook.OpenOptions{
				NotebookIdentifier: args[0],
			}

			if err := notebook.OpenNotebook(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}
}
