package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

func getRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename [current-name-or-id] [new-name]",
		Short: "重命名笔记本",
		Long: `重命名现有的思源笔记笔记本。

参数：
- current-name-or-id: 当前笔记本的名称或ID（必需）
- new-name: 新的笔记本名称（必需）

智能匹配规则：
- 如果输入符合ID格式（如 20231201120000-abc123），直接使用ID
- 否则进行名称匹配：精确匹配 → 不区分大小写匹配 → 包含匹配
- 如果找到多个匹配项，会列出所有匹配的笔记本供参考

新名称要求：
- 不能为空
- 长度不超过50个字符
- 不能与现有笔记本重名（不区分大小写）
- 建议使用简洁明了的名称

示例：
- siyuan-cli notebook rename "旧笔记本" "新笔记本"
- siyuan-cli notebook rename 20231201120000-abc123 "重命名后的笔记本"
- siyuan-cli notebook list  # 查看所有笔记本`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println("❌ 错误: 请提供当前笔记本名称/ID和新名称")
				fmt.Println("💡 使用方法: siyuan-cli notebook rename <当前名称或ID> <新名称>")
				fmt.Println("💡 示例: siyuan-cli notebook rename \"旧笔记本\" \"新笔记本\"")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				return
			}

			opts := notebook.RenameOptions{
				CurrentIdentifier: args[0],
				NewName:           args[1],
			}

			if err := notebook.RenameNotebook(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}
}
