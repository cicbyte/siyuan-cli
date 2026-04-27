/*
Copyright © 2025 cicbyte
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/cicbyte/siyuan-cli/cmd/auth"
	"github.com/cicbyte/siyuan-cli/cmd/block"
	"github.com/cicbyte/siyuan-cli/cmd/chat"
	"github.com/cicbyte/siyuan-cli/cmd/config"
	"github.com/cicbyte/siyuan-cli/cmd/document"
	"github.com/cicbyte/siyuan-cli/cmd/export"
	"github.com/cicbyte/siyuan-cli/cmd/fav"
	"github.com/cicbyte/siyuan-cli/cmd/importcmd"
	"github.com/cicbyte/siyuan-cli/cmd/mcp"
	"github.com/cicbyte/siyuan-cli/cmd/notebook"
	"github.com/cicbyte/siyuan-cli/cmd/asset"
	"github.com/cicbyte/siyuan-cli/cmd/synccmd"
	"github.com/cicbyte/siyuan-cli/cmd/query"
	"github.com/cicbyte/siyuan-cli/cmd/search"
	"github.com/cicbyte/siyuan-cli/cmd/tag"
	"github.com/cicbyte/siyuan-cli/cmd/version"
	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/utils"
	"github.com/spf13/cobra"
)

var globalFormat string

var rootCmd = &cobra.Command{
	Use:   "siyuan-cli",
	Short: "思源笔记命令行工具",
	Long: `思源笔记命令行工具 (SiYuan CLI)

这是一个功能强大的思源笔记命令行管理工具，支持：
- 笔记本管理 (notebook)
- 文档操作 (document)
- 搜索与查询 (search / query)
- 块操作 (block)
- 标签管理 (tag)
- 导入导出 (export / import)
- 资源管理 (asset)
- 同步管理 (sync)
- 收藏功能 (fav)
- 配置管理 (config)
- AI 对话 (chat)
- MCP Server (mcp)

开始使用：
  siyuan-cli notebook list    # 列出笔记本
  siyuan-cli document list    # 列出文档
  siyuan-cli search block "关键词"  # 全文搜索
  siyuan-cli query "SELECT * FROM blocks WHERE type='d' LIMIT 10"
  siyuan-cli chat "搜索关于 Go 的笔记"  # AI 对话

更多帮助：
  siyuan-cli --help
  siyuan-cli [command] --help`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 初始化应用目录
	if err := utils.InitAppDirs(); err != nil {
		fmt.Printf("初始化目录失败: %v\n", err)
		os.Exit(1)
	}
	// 加载配置(会自动创建默认配置)
	common.SetAppConfig(utils.ConfigInstance.LoadConfig())
	// 初始化日志
	if err := log.Init(utils.ConfigInstance.GetLogPath()); err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 注册命令模块
	rootCmd.AddCommand(auth.GetAuthCommand())
	rootCmd.AddCommand(chat.GetChatCommand())
	rootCmd.AddCommand(notebook.GetNotebookCommand())
	rootCmd.AddCommand(document.GetDocumentCommand())
	rootCmd.AddCommand(fav.GetFavCommand())
	rootCmd.AddCommand(config.GetConfigCommand())
	rootCmd.AddCommand(search.GetSearchCommand())
	rootCmd.AddCommand(query.GetQueryCommand())
	rootCmd.AddCommand(block.GetBlockCommand())
	rootCmd.AddCommand(tag.GetTagCommand())
	rootCmd.AddCommand(export.GetExportCommand())
	rootCmd.AddCommand(importcmd.GetImportCommand())
	rootCmd.AddCommand(asset.GetAssetCommand())
	rootCmd.AddCommand(synccmd.GetSyncCommand())
	rootCmd.AddCommand(version.GetVersionCommand())
	rootCmd.AddCommand(mcp.GetMCPCommand())

	rootCmd.PersistentFlags().StringVarP(&globalFormat, "format", "f", "", "输出格式 (table|json)")

	// 认证检查：未配置思源连接时引导用户
	skipCommands := map[string]bool{"auth": true, "chat": true, "config": true, "version": true, "help": true, "completion": true, "mcp": true}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// 设置全局输出格式
		if globalFormat == "" {
			globalFormat = common.GetAppConfig().Output.Format
			if globalFormat == "" {
				globalFormat = "table"
			}
		}
		output.SetFormat(globalFormat)

		if skipCommands[cmd.Name()] || (cmd.HasParent() && skipCommands[cmd.Parent().Name()]) {
			return nil
		}
		cfg := common.GetAppConfig()
		if !cfg.SiYuan.Enabled || cfg.SiYuan.BaseURL == "" {
			fmt.Fprintln(os.Stderr, "欢迎使用 siyuan-cli！")
			fmt.Fprintln(os.Stderr, "  尚未配置思源笔记连接，请先运行：")
			fmt.Fprintln(os.Stderr, "  siyuan-cli auth login")
			fmt.Fprintln(os.Stderr, "  登录时会交互式引导配置思源笔记地址和 API Token")
			os.Exit(1)
		}
		return nil
	}
}
