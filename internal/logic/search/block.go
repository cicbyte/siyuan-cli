package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

type SearchBlockOptions struct {
	Keyword  string
	Notebook string
	Limit    int
	OutputFile string
}

func SearchBlock(opts SearchBlockOptions) error {
	logger := log.GetLogger()
	logger.Info("搜索块", zap.String("keyword", opts.Keyword))

	if strings.TrimSpace(opts.Keyword) == "" {
		fmt.Println("❌ 错误: 请提供搜索关键词")
		fmt.Println("💡 使用方法: siyuan-cli search block <关键词>")
		return fmt.Errorf("搜索关键词不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := client.FullTextSearchBlock(ctx, siyuan.FullTextSearchBlockOptions{Query: opts.Keyword})
	if err != nil {
		fmt.Printf("❌ 搜索块失败: %v\n", err)
		return err
	}

	if opts.Notebook != "" {
		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
			return err
		}
		nbID, _, err := document.FindNotebook(notebooks, opts.Notebook)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}
		filtered := make([]siyuan.FullTextSearchBlockResult, 0)
		for _, r := range results {
			if r.Box == nbID {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(results, opts.OutputFile)
	}

	if len(results) == 0 {
		fmt.Println("未找到匹配的块")
		return nil
	}

	fmt.Printf("找到 %d 个匹配块:\n\n", len(results))
	headers := []string{"ID", "类型", "内容", "文档路径"}
	rows := make([][]string, 0, len(results))
	for _, r := range results {
		content := strings.ReplaceAll(r.Content, "\n", " ")
		rows = append(rows, []string{
			r.ID,
			r.BlockType,
			output.Truncate(content, 50),
			r.HPath,
		})
	}
	output.PrintTable(headers, rows)
	return nil
}
