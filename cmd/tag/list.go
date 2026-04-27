package tag

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/tag"
	"github.com/spf13/cobra"
)

var tagListOutput string

func getListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有标签",
		Long:  "列出思源笔记中的所有标签。",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tag.ListTags(tag.ListOptions{
				OutputFile: tagListOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&tagListOutput, "output", "o", "", "导出到文件")
	return cmd
}

var tagSearchOutput string

func getSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <keyword>",
		Short: "搜索标签",
		Long:  "搜索匹配关键词的标签。",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := tag.SearchTags(tag.SearchOptions{
				Keyword:    args[0],
				OutputFile: tagSearchOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&tagSearchOutput, "output", "o", "", "导出到文件")
	return cmd
}
