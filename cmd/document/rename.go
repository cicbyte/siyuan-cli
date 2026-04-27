package document

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

func getRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename [notebook-name-or-id] [document-path] [new-name]",
		Short: "重命名文档",
		Long: `重命名指定的思源笔记文档。

参数：
- notebook-name-or-id: 目标笔记本名称或ID（必需）
- document-path: 文档路径（必需）
- new-name: 新文档名称（必需）


智能匹配规则：
- 如果输入符合ID格式（如 20231201120000-abc123），直接使用ID
- 否则进行名称匹配：精确匹配 → 不区分大小写匹配 → 包含匹配
- 如果找到多个匹配项，会列出所有匹配的笔记本供参考

文档路径格式：
- 必须以 / 开头，如 "/项目文档/API文档"
- 路径中包含文档的完整人类可读路径

新文档名称：
- 不能包含特殊字符
- 建议使用简短、有意义的名称

- siyuan-cli notebook list  # 查看所有笔记本
- siyuan-cli document ls <笔记本>  # 查看文档结构

示例：
- siyuan-cli document rename python "/旧API文档" "新API文档"
- siyuan-cli document rename "工作笔记" "/项目/README" "项目说明"`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 3 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID、文档路径和新名称")
				fmt.Println("💡 使用方法: siyuan-cli document rename <笔记本名称或ID> <文档路径> <新名称>")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				fmt.Println("💡 使用 'siyuan-cli document ls <笔记本>' 查看文档结构")
				return
			}

			opts := document.RenameOptions{
				NotebookIdentifier: args[0],
				DocumentPath:       args[1],
				NewName:            args[2],
			}

			if err := document.RenameDocument(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}
}
