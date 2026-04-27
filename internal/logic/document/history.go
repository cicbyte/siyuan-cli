package document

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

type HistoryOptions struct {
	NotebookIdentifier string
	Path               string
	Query              string
	Page               int
	OutputFile         string
}

func GetDocumentHistory(opts HistoryOptions) error {
	logger := log.GetLogger()
	logger.Info("获取文档历史", zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path))

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var notebook string
	if opts.NotebookIdentifier != "" {
		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
			return err
		}
		_, _, err = FindNotebook(notebooks, opts.NotebookIdentifier)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}
		notebook = opts.NotebookIdentifier
	}

	query := opts.Query
	if query == "" && opts.Path != "" {
		query = opts.Path
	}

	searchResult, err := client.SearchHistory(ctx, siyuan.SearchHistoryOptions{
		Query:    query,
		Notebook: notebook,
		Page:     opts.Page,
	})
	if err != nil {
		fmt.Printf("❌ 获取文档历史失败: %v\n", err)
		return err
	}

	if len(searchResult.Histories) == 0 {
		if output.IsJSON("") || opts.OutputFile != "" {
			data := map[string]any{
				"query":      query,
				"notebook":   notebook,
				"pageCount":  searchResult.PageCount,
				"totalCount": searchResult.TotalCount,
				"history":    []any{},
			}
			if opts.OutputFile != "" {
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				os.WriteFile(opts.OutputFile, jsonData, 0644)
				fmt.Printf("已导出到: %s\n", opts.OutputFile)
				return nil
			}
			output.PrintJSON(data)
			return nil
		}
		fmt.Println("没有找到历史记录")
		return nil
	}

	// 获取第一页的详细条目
	created := searchResult.Histories[0]
	items, err := client.GetHistoryItems(ctx, siyuan.GetHistoryItemsOptions{
		Created: created,
		Query:   query,
		Notebook: notebook,
	})
	if err != nil {
		fmt.Printf("❌ 获取历史详情失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		data := map[string]any{
			"query":      query,
			"notebook":   notebook,
			"pageCount":  searchResult.PageCount,
			"totalCount": searchResult.TotalCount,
			"count":      len(items),
			"history":    items,
		}
		if opts.OutputFile != "" {
			jsonData, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("序列化JSON失败: %w", err)
			}
			dir := filepath.Dir(opts.OutputFile)
			if dir != "." {
				os.MkdirAll(dir, 0755)
			}
			if err := os.WriteFile(opts.OutputFile, jsonData, 0644); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}
			fmt.Printf("已导出到: %s\n", opts.OutputFile)
			return nil
		}
		output.PrintJSON(data)
		return nil
	}

	if len(items) == 0 {
		fmt.Println("没有找到历史记录")
		return nil
	}

	fmt.Printf("文档历史 (共 %d 条, 第 %d/%d 页, 时间: %s):\n\n", searchResult.TotalCount, opts.Page, searchResult.PageCount, created)
	headers := []string{"标题", "操作", "文档路径", "笔记本"}
	rows := make([][]string, 0, len(items))
	for _, h := range items {
		rows = append(rows, []string{h.Title, h.Op, h.Path, h.Notebook})
	}
	output.PrintTable(headers, rows)
	return nil
}

type RollbackOptions struct {
	NotebookIdentifier string
	HistoryPath        string
}

func RollbackDocument(opts RollbackOptions) error {
	logger := log.GetLogger()
	logger.Info("回滚文档", zap.String("notebook", opts.NotebookIdentifier), zap.String("history_path", opts.HistoryPath))

	if strings.TrimSpace(opts.HistoryPath) == "" {
		fmt.Println("❌ 错误: 历史路径不能为空")
		fmt.Println("💡 使用方法: siyuan-cli document rollback --notebook <笔记本> --to <历史路径>")
		return fmt.Errorf("历史路径不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var notebookID string
	if opts.NotebookIdentifier != "" {
		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
			return err
		}
		nbID, _, err := FindNotebook(notebooks, opts.NotebookIdentifier)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}
		notebookID = nbID
	}

	err = client.RollbackDocHistory(ctx, notebookID, opts.HistoryPath)
	if err != nil {
		fmt.Printf("❌ 回滚失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 文档已回滚到历史版本 %s\n", opts.HistoryPath)
	return nil
}
