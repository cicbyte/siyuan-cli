package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

var notebookIcon string

func getCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [notebook-name]",
		Short: "创建笔记本",
		Long: `创建新的思源笔记笔记本。

参数：
- notebook-name: 新笔记本的名称（必需）

选项：
- --icon, -i: 设置笔记本图标（可选）

笔记本名称要求：
- 不能为空
- 长度不超过50个字符
- 不能与现有笔记本重名（不区分大小写）
- 建议使用简洁明了的名称

图标要求：
- 可选参数，不设置则使用默认图标
- 长度不超过10个字符
- 支持emoji或文本图标

示例：
- siyuan-cli notebook create "我的笔记本"
- siyuan-cli notebook create "工作笔记" --icon "💼"
- siyuan-cli notebook create "学习资料" -i "📚"`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称")
				fmt.Println("💡 使用方法: siyuan-cli notebook create <笔记本名称>")
				fmt.Println("💡 示例: siyuan-cli notebook create \"我的笔记本\"")
				return
			}

			opts := notebook.CreateOptions{
				Name: args[0],
				Icon: notebookIcon,
			}

			if err := notebook.CreateNotebook(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&notebookIcon, "icon", "i", "", "设置笔记本图标")
	return cmd
}
