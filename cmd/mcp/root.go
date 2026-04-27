package mcp

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetMCPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "启动 MCP Server",
		Long: `以 stdio 模式启动 MCP Server，让 AI 客户端能直接搜索和操作思源笔记数据。

注册的 Tools:
  notebook_list       列出所有笔记本
  document_list       列出文档树
  document_get        获取文档内容
  document_outline    获取文档大纲
  block_get           获取块信息
  block_get_kramdown  获取块 kramdown 源码
  search_fulltext     全文搜索块
  search_docs         搜索文档
  tag_list            列出所有标签
  query_sql           执行 SQL 查询
  document_create     创建文档
  daily_note_create   创建日记
  block_update        更新块内容
  block_append        追加块到文档`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runMCPServer(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			return nil
		},
	}
}
