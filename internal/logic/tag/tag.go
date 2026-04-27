package tag

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

type ListOptions struct {
	OutputFile string
}

func ListTags(opts ListOptions) error {
	logger := log.GetLogger()
	logger.Info("列出所有标签")

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tagTree, err := client.GetTag(ctx)
	if err != nil {
		fmt.Printf("❌ 获取标签列表失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(tagTree, opts.OutputFile)
	}

	labels := siyuan.FlattenTags(tagTree)
	if len(labels) == 0 {
		fmt.Println("暂无标签")
		return nil
	}

	fmt.Printf("共 %d 个标签:\n\n", len(labels))
	for _, label := range labels {
		fmt.Printf("  #%s\n", label)
	}
	return nil
}

type SearchOptions struct {
	Keyword    string
	OutputFile string
}

func SearchTags(opts SearchOptions) error {
	logger := log.GetLogger()
	logger.Info("搜索标签", zap.String("keyword", opts.Keyword))

	if strings.TrimSpace(opts.Keyword) == "" {
		fmt.Println("❌ 错误: 请提供搜索关键词")
		fmt.Println("💡 使用方法: siyuan-cli tag search <关键词>")
		return fmt.Errorf("搜索关键词不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	labels, err := client.SearchTag(ctx, opts.Keyword)
	if err != nil {
		fmt.Printf("❌ 搜索标签失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(labels, opts.OutputFile)
	}

	if len(labels) == 0 {
		fmt.Println("未找到匹配的标签")
		return nil
	}

	fmt.Printf("找到 %d 个匹配标签:\n\n", len(labels))
	for _, label := range labels {
		fmt.Printf("  #%s\n", label)
	}
	return nil
}

type AddTagOptions struct {
	DocID              string
	NotebookIdentifier string
	Path               string
	Tags               []string
}

func AddTags(opts AddTagOptions) error {
	logger := log.GetLogger()
	logger.Info("添加标签", zap.String("docID", opts.DocID), zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path), zap.Strings("tags", opts.Tags))

	if len(opts.Tags) == 0 {
		fmt.Println("❌ 错误: 请使用 --tag 指定至少一个标签")
		fmt.Println("💡 使用方法: siyuan-cli tag add <doc-id> --tag \"标签1\" --tag \"标签2\"")
		return fmt.Errorf("标签不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var docID string
	if opts.DocID != "" {
		docID = opts.DocID
	} else {
		resolved, err := resolveDocID(ctx, client, opts.NotebookIdentifier, opts.Path)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}
		docID = resolved
	}

	existing := make(map[string]string)
	attrs, err := client.GetBlockAttrs(ctx, docID)
	if err == nil && attrs != nil && attrs.Attrs != nil {
		if tags, ok := attrs.Attrs["tags"]; ok && tags != "" {
			for _, t := range strings.Split(tags, ",") {
				existing[strings.TrimSpace(t)] = ""
			}
		}
	}

	for _, tag := range opts.Tags {
		existing[tag] = ""
	}

	tagList := make([]string, 0, len(existing))
	for t := range existing {
		tagList = append(tagList, t)
	}

	err = client.SetBlockAttrs(ctx, siyuan.SetBlockAttrsOptions{
		ID:    docID,
		Attrs: map[string]string{"tags": strings.Join(tagList, ",")},
	})
	if err != nil {
		fmt.Printf("❌ 设置标签失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已为文档设置标签: %s\n", strings.Join(opts.Tags, ", "))
	return nil
}

type RemoveTagOptions struct {
	DocID              string
	NotebookIdentifier string
	Path               string
	Tag                string
}

func RemoveTag(opts RemoveTagOptions) error {
	logger := log.GetLogger()
	logger.Info("移除标签", zap.String("docID", opts.DocID), zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path), zap.String("tag", opts.Tag))

	if strings.TrimSpace(opts.Tag) == "" {
		fmt.Println("❌ 错误: 请使用 --tag 指定要移除的标签")
		fmt.Println("💡 使用方法: siyuan-cli tag remove <doc-id> --tag \"标签\"")
		return fmt.Errorf("标签不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var docID string
	if opts.DocID != "" {
		docID = opts.DocID
	} else {
		resolved, err := resolveDocID(ctx, client, opts.NotebookIdentifier, opts.Path)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}
		docID = resolved
	}

	attrs, err := client.GetBlockAttrs(ctx, docID)
	if err != nil {
		fmt.Printf("❌ 获取文档属性失败: %v\n", err)
		return err
	}

	existing := make(map[string]string)
	if attrs != nil && attrs.Attrs != nil {
		if tags, ok := attrs.Attrs["tags"]; ok && tags != "" {
			for _, t := range strings.Split(tags, ",") {
				existing[strings.TrimSpace(t)] = ""
			}
		}
	}

	delete(existing, opts.Tag)

	tagList := make([]string, 0, len(existing))
	for t := range existing {
		tagList = append(tagList, t)
	}

	err = client.SetBlockAttrs(ctx, siyuan.SetBlockAttrsOptions{
		ID:    docID,
		Attrs: map[string]string{"tags": strings.Join(tagList, ",")},
	})
	if err != nil {
		fmt.Printf("❌ 设置标签失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已移除标签: %s\n", opts.Tag)
	return nil
}

func resolveDocID(ctx context.Context, client *siyuan.Client, notebookID, path string) (string, error) {
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	targetID, targetName, err := document.FindNotebook(notebooks, notebookID)
	if err != nil {
		return "", err
	}

	resolvedPath, err := document.ResolveDocPath(ctx, client, targetID, targetName, path)
	if err != nil {
		return "", err
	}

	docID := strings.TrimSuffix(strings.Trim(resolvedPath, "/"), "/")
	parts := strings.Split(docID, "/")
	return parts[len(parts)-1], nil
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
