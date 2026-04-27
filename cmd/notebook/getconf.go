package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

func getGetConfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getconf [notebook-name-or-id]",
		Short: "获取笔记本配置",
		Long: `获取指定思源笔记笔记本的详细配置信息，支持名称或ID。

参数：
	- notebook-name-or-id: 要获取配置的笔记本名称或ID（必需）

选项：
	- --format, -f: 输出格式 (table|json)，全局默认为 table
	- --output, -o: 导出到指定文件（自动使用JSON格式）

配置信息包括：
	- 基本信息：名称、ID、图标、状态
	- 排序设置：排序顺序、排序模式
	- 时间信息：创建时间、更新时间
	- 高级配置：头像、排序算法、引用设置等

示例：
	- siyuan-cli notebook getconf "测试笔记本"                    # 使用名称，table显示
	- siyuan-cli notebook getconf 20251105164527-ezlspgg          # 使用ID
	- siyuan-cli notebook getconf "工作笔记" --format json        # JSON格式输出
	- siyuan-cli notebook getconf "项目笔记" -o config.json       # 导出到JSON文件
	- siyuan-cli notebook list  # 查看所有笔记本`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID")
				fmt.Println("💡 使用方法: siyuan-cli notebook getconf <笔记本名称或ID>")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				return
			}

			opts := notebook.GetConfOptions{
				NotebookIdentifier: args[0],
				OutputFile:         outputFile,
			}

			if err := notebook.GetNotebookConf(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "导出到指定文件")

	return cmd
}
