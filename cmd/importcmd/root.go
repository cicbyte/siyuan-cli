package importcmd

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/importcmd"
	"github.com/spf13/cobra"
)

var importMdNotebook string
var importMdPath string
var importSyNotebook string

func GetImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "导入文档",
		Long: `导入 Markdown 或思源格式文档。

支持格式：
- md: 标准 Markdown 文件或目录
- sy: 思源格式数据`,
	}
	cmd.AddCommand(getMdCommand())
	cmd.AddCommand(getSyCommand())
	return cmd
}

func getMdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "md <file-or-dir>",
		Short: "导入 Markdown",
		Long: `导入 Markdown 文件或目录到指定笔记本。

示例：
  siyuan-cli import md ./notes/ --notebook java --path "导入"
  siyuan-cli import md article.md --notebook blog`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := importcmd.ImportMd(importcmd.ImportMdOptions{
				File:               args[0],
				NotebookIdentifier: importMdNotebook,
				DestPath:           importMdPath,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&importMdNotebook, "notebook", "n", "", "目标笔记本名称或ID")
	cmd.Flags().StringVarP(&importMdPath, "path", "p", "", "目标路径")
	_ = cmd.MarkFlagRequired("notebook")
	return cmd
}

func getSyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sy <file-or-dir>",
		Short: "导入思源格式",
		Long: `导入思源格式数据到指定笔记本。

示例：
  siyuan-cli import sy ./data/ --notebook java`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := importcmd.ImportSy(importcmd.ImportSyOptions{
				File:               args[0],
				NotebookIdentifier: importSyNotebook,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&importSyNotebook, "notebook", "n", "", "目标笔记本名称或ID")
	_ = cmd.MarkFlagRequired("notebook")
	return cmd
}
