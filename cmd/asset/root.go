package asset

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/assetcmd"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/spf13/cobra"
)

func GetAssetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "资源管理",
		Long: `思源笔记资源文件管理工具。

可以用来：
- 上传资源文件
- 查看文档关联的资源
- 列出未使用资源
- 清理未使用资源`,
	}
	cmd.AddCommand(getUploadCommand())
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getUnusedCommand())
	cmd.AddCommand(getCleanCommand())
	return cmd
}

func getUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "上传资源文件",
		Long:  "上传文件到思源笔记资源目录。",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := assetcmd.UploadAsset(assetcmd.UploadOptions{
				FilePath: args[0],
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	return cmd
}

var assetListOutput string

func getListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [doc-id | notebook] [doc-path]",
		Short: "查看文档资源",
		Long: `查看文档关联的资源文件列表。支持直接传文档 ID 或 笔记本/文档路径。

示例：
  siyuan-cli asset list 20240101120000-xxx
  siyuan-cli asset list java "笔记/重要"`,
		Args: cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var opts assetcmd.ListOptions
			opts.OutputFile = assetListOutput
			if len(args) == 0 {
				// 不指定文档，列出最近上传的资源
			} else if len(args) == 1 && document.IsDocID(args[0]) {
				opts.DocID = args[0]
			} else if len(args) == 2 {
				opts.NotebookIdentifier = args[0]
				opts.Path = args[1]
			} else {
				fmt.Println("❌ 错误: 参数格式不正确")
				fmt.Println("💡 使用方法: siyuan-cli asset list [doc-id | notebook doc-path]")
				return
			}
			if err := assetcmd.ListAssets(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&assetListOutput, "output", "o", "", "导出到文件")
	return cmd
}

var assetUnusedOutput string

func getUnusedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unused",
		Short: "列出未使用资源",
		Long:  "列出未被任何文档引用的资源文件。",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := assetcmd.ListUnusedAssets(assetcmd.UnusedOptions{
				OutputFile: assetUnusedOutput,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&assetUnusedOutput, "output", "o", "", "导出到文件")
	return cmd
}

var assetCleanForce bool

func getCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "清理未使用资源",
		Long:  "清理未被任何文档引用的资源文件。",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := assetcmd.CleanUnusedAssets(assetcmd.CleanOptions{
				Force: assetCleanForce,
			}); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}
	cmd.Flags().BoolVar(&assetCleanForce, "force", false, "强制清理（跳过确认）")
	return cmd
}
