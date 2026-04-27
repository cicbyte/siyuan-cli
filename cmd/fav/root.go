package fav

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/fav"
	"github.com/spf13/cobra"
)

func GetFavCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fav [content]",
		Short: "收藏内容到思源笔记",
		Long: `将内容收藏到"我的收藏"笔记本中，按年月自动分类存储。

支持多种输入方式：
- 直接提供内容作为参数
- 通过管道输入内容
- 自动提取标题或使用时间戳

自动组织结构：
- 笔记本: 我的收藏
- 年份目录: /2025, /2026 等
- 月份目录: /01, /02 等

智能标题处理：
- 如果内容以 # 开头，自动提取为文档标题
- 否则使用时间戳作为文档标题

使用示例:
  siyuan-cli fav "这是一个重要的代码片段"
  echo "# 学习笔记\n\n今天学习了Go语言的基础知识" | siyuan-cli fav
  cat README.md | siyuan-cli fav`,
		Run: func(cmd *cobra.Command, args []string) {
			var content string

			if len(args) > 0 {
				content = args[0]
			}

			opts := fav.FavOptions{
				Content: content,
			}

			if err := fav.AddToFavorites(opts); err != nil {
				fmt.Printf("❌ 收藏失败: %v\n", err)
				return
			}
		},
	}
}
