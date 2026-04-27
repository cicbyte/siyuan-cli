package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

type SearchDocOptions struct {
	Keyword    string
	Notebook   string
	Limit      int
	OutputFile string
}

func SearchDoc(opts SearchDocOptions) error {
	logger := log.GetLogger()
	logger.Info("搜索文档", zap.String("keyword", opts.Keyword), zap.String("notebook", opts.Notebook))

	if strings.TrimSpace(opts.Keyword) == "" {
		fmt.Println("❌ 错误: 请提供搜索关键词")
		fmt.Println("💡 使用方法: siyuan-cli search doc <关键词>")
		return fmt.Errorf("搜索关键词不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := client.SearchDocs(ctx, siyuan.SearchDocsOptions{K: opts.Keyword})
	if err != nil {
		fmt.Printf("❌ 搜索文档失败: %v\n", err)
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
		filtered := make([]siyuan.SearchDocResult, 0)
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
		type enrichedResult struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			HPath string `json:"hpath"`
			Box   string `json:"box"`
			Path  string `json:"path"`
		}
		enriched := make([]enrichedResult, len(results))
		for i, r := range results {
			enriched[i] = enrichedResult{
				ID:    docIDFromPath(r.Path),
				Title: docTitle(r.Title, r.HPath),
				HPath: r.HPath,
				Box:   r.Box,
				Path:  r.Path,
			}
		}
		return outputJSON(enriched, opts.OutputFile)
	}

	if len(results) == 0 {
		fmt.Println("未找到匹配的文档")
		return nil
	}

	fmt.Printf("找到 %d 个匹配文档:\n\n", len(results))
	headers := []string{"ID", "标题", "路径"}
	rows := make([][]string, 0, len(results))
	for _, r := range results {
		id := r.ID
		if id == "" {
			id = docIDFromPath(r.Path)
		}
		rows = append(rows, []string{
			id,
			output.Truncate(docTitle(r.Title, r.HPath), 40),
			r.HPath,
		})
	}
	output.PrintTable(headers, rows)
	return nil
}

func docIDFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".sy")
}

func docTitle(title, hpath string) string {
	if title != "" {
		return title
	}
	// 从 hpath 最后一段提取标题（如 "编程/读书笔记/三体" → "三体"）
	if idx := strings.LastIndex(hpath, "/"); idx >= 0 {
		return hpath[idx+1:]
	}
	return hpath
}

func outputJSON(data any, outputFile string) error {
	if outputFile != "" {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化JSON失败: %w", err)
		}
		dir := filepath.Dir(outputFile)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建输出目录失败: %w", err)
			}
		}
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
		fmt.Printf("已导出到: %s\n", outputFile)
		return nil
	}
	output.PrintJSON(data)
	return nil
}
