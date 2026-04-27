package notebook

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/notebook"
	"github.com/spf13/cobra"
)

var (
	setConfMode     string
	setConfFile     string
	setConfName     string
	setConfIcon     string
	setConfSort     string
	setConfSortMode string
	setConfClosed   string
)

func getSetConfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setconf [notebook-name-or-id]",
		Short: "设置笔记本配置",
		Long: `设置指定思源笔记笔记本的配置信息，支持名称或ID。

参数：
- notebook-name-or-id: 要设置配置的笔记本名称或ID（必需）

选项：
- --mode, -m: 输入模式 (cli|json|tui)，默认为 tui
- --input, -i: JSON配置文件路径（配合json模式使用）
- --name, -n: 设置笔记本名称（cli模式）
- --icon: 设置笔记本图标（cli模式）
- --sort: 设置排序顺序（cli模式）
- --sort-mode: 设置排序模式（cli模式）
- --closed: 设置是否关闭（cli模式）

输入模式：
- cli: 命令行直接设置，适合脚本和自动化
- json: 从JSON文件导入配置
- tui: 交互式界面，可视化设置（推荐）

支持配置项：
- 基本信息：名称、图标、状态
- 排序设置：排序顺序、排序模式
- 高级配置：引用创建锚点、文档保存文件夹

CLI模式示例：
- siyuan-cli notebook setconf "工作笔记" --name "新工作笔记"
- siyuan-cli notebook setconf "项目笔记" --icon "📚" --sort 10

JSON模式示例：
- siyuan-cli notebook setconf "配置笔记" --mode json --input config.json

TUI模式示例：
- siyuan-cli notebook setconf "个人笔记" --mode tui
- siyuan-cli notebook setconf "设置笔记"  # 默认TUI模式

注意：修改配置可能需要重启思源笔记才能完全生效`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ 错误: 请提供笔记本名称或ID")
				fmt.Println("💡 使用方法: siyuan-cli notebook setconf <笔记本名称或ID>")
				fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
				return
			}

			configData := make(map[string]interface{})

			if cmd.Flags().Changed("name") {
				configData["name"] = cmd.Flag("name").Value.String()
			}
			if cmd.Flags().Changed("icon") {
				configData["icon"] = cmd.Flag("icon").Value.String()
			}
			if cmd.Flags().Changed("sort") {
				configData["sort"] = cmd.Flag("sort").Value.String()
			}
			if cmd.Flags().Changed("sort-mode") {
				configData["sortMode"] = cmd.Flag("sort-mode").Value.String()
			}
			if cmd.Flags().Changed("closed") {
				configData["closed"] = cmd.Flag("closed").Value.String()
			}

			opts := notebook.SetConfOptions{
				NotebookIdentifier: args[0],
				InputMode:          setConfMode,
				ConfigFile:         setConfFile,
				ConfigData:         configData,
			}

			if err := notebook.SetNotebookConf(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&setConfMode, "mode", "m", "tui", "输入模式 (cli|json|tui)")
	cmd.Flags().StringVarP(&setConfFile, "input", "i", "", "JSON配置文件路径")
	cmd.Flags().StringVarP(&setConfName, "name", "n", "", "设置笔记本名称")
	cmd.Flags().StringVar(&setConfIcon, "icon", "", "设置笔记本图标")
	cmd.Flags().StringVar(&setConfSort, "sort", "", "设置排序顺序")
	cmd.Flags().StringVar(&setConfSortMode, "sort-mode", "", "设置排序模式")
	cmd.Flags().StringVar(&setConfClosed, "closed", "", "设置是否关闭")

	cmd.RegisterFlagCompletionFunc("mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"cli", "json", "tui"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
