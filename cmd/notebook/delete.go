package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

var forceDelete bool

func getDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [notebook-name-or-id]",
		Short: "删除笔记本",
		Long: `删除指定的思源笔记笔记本，支持名称或ID。

参数：
- notebook-name-or-id: 要删除的笔记本名称或ID（必需）

选项：
- --force, -f: 强制删除，跳过确认提示

⚠️  危险操作警告：
- 此操作不可逆，笔记本中的所有文档和数据将被永久删除
- 删除前请确保已经备份重要数据
- 建议先关闭笔记本再进行删除操作

智能匹配规则：
- 如果输入符合ID格式（如 20231201120000-abc123），直接使用ID
- 否则进行名称匹配：精确匹配 → 不区分大小写匹配 → 包含匹配
- 如果找到多个匹配项，会列出所有匹配的笔记本供参考

安全建议：
- 删除前可使用 'siyuan-cli notebook list' 确认笔记本信息
- 对于重要笔记本，建议先导出备份再删除
- 使用 --force 选项时要特别谨慎

示例：
- siyuan-cli notebook delete "测试笔记本"                    # 使用名称
- siyuan-cli notebook delete 20251105164527-ezlspgg          # 使用ID
- siyuan-cli notebook delete "临时笔记" --force              # 强制删除
- siyuan-cli notebook list  # 查看所有笔记本`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID")
				fmt.Println("💡 使用方法: siyuan-cli notebook delete <笔记本名称或ID>")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				fmt.Println("⚠️  注意：此操作不可逆，请谨慎操作")
				return
			}

			opts := notebook.DeleteOptions{
				NotebookIdentifier: args[0],
				Force:              forceDelete,
			}

			if err := notebook.DeleteNotebook(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().BoolVarP(&forceDelete, "force", "F", false, "强制删除，跳过确认提示")
	return cmd
}
