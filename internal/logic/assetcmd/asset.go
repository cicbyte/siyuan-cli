package assetcmd

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

type UploadOptions struct {
	FilePath string
}

func UploadAsset(opts UploadOptions) error {
	logger := log.GetLogger()
	logger.Info("上传资源文件", zap.String("file", opts.FilePath))

	if strings.TrimSpace(opts.FilePath) == "" {
		fmt.Println("❌ 错误: 请提供文件路径")
		fmt.Println("💡 使用方法: siyuan-cli asset upload <文件路径>")
		return fmt.Errorf("文件路径不能为空")
	}

	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		fmt.Printf("❌ 文件不存在: %s\n", opts.FilePath)
		return err
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := client.UploadAsset(ctx, opts.FilePath)
	if err != nil {
		fmt.Printf("❌ 上传失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 文件已上传: %s\n", result)
	return nil
}

type ListOptions struct {
	DocID              string
	NotebookIdentifier string
	Path               string
	OutputFile         string
}

func ListAssets(opts ListOptions) error {
	logger := log.GetLogger()
	logger.Info("列出文档资源", zap.String("docID", opts.DocID), zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path))

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
	} else if opts.NotebookIdentifier != "" && opts.Path != "" {
		resolved, err := resolveDocID(ctx, client, opts.NotebookIdentifier, opts.Path)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}
		docID = resolved
	} else {
		// 无参数模式：提示用法
		fmt.Println("💡 请指定文档以查看其关联资源")
		fmt.Println("   siyuan-cli asset list <doc-id>")
		fmt.Println("   siyuan-cli asset list <笔记本> <文档路径>")
		fmt.Println("   siyuan-cli asset unused  # 查看未使用资源")
		return nil
	}

	assets, err := client.GetDocAssets(ctx, docID)
	if err != nil {
		fmt.Printf("❌ 获取文档资源失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(assets, opts.OutputFile)
	}

	if len(assets) == 0 {
		fmt.Println("该文档没有关联资源文件")
		return nil
	}

	fmt.Printf("共 %d 个资源文件:\n\n", len(assets))
	headers := []string{"名称", "类型", "大小", "更新时间"}
	rows := make([][]string, 0, len(assets))
	for _, a := range assets {
		size := formatSize(a.Size)
		rows = append(rows, []string{a.Name, a.Type, size, a.Updated})
	}
	output.PrintTable(headers, rows)
	return nil
}

type UnusedOptions struct {
	OutputFile string
}

func ListUnusedAssets(opts UnusedOptions) error {
	logger := log.GetLogger()
	logger.Info("列出未使用资源")

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	assets, err := client.GetUnusedAssets(ctx)
	if err != nil {
		fmt.Printf("❌ 获取未使用资源失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(assets, opts.OutputFile)
	}

	if len(assets) == 0 {
		fmt.Println("没有未使用的资源文件")
		return nil
	}

	fmt.Printf("共 %d 个未使用资源文件:\n\n", len(assets))
	headers := []string{"路径", "大小", "更新时间"}
	rows := make([][]string, 0, len(assets))
	for _, a := range assets {
		size := formatSize(a.Size)
		rows = append(rows, []string{a.Path, size, a.Updated})
	}
	output.PrintTable(headers, rows)
	return nil
}

type CleanOptions struct {
	Force bool
}

func CleanUnusedAssets(opts CleanOptions) error {
	logger := log.GetLogger()
	logger.Info("清理未使用资源", zap.Bool("force", opts.Force))

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = client.RemoveUnusedAssets(ctx)
	if err != nil {
		fmt.Printf("❌ 清理失败: %v\n", err)
		return err
	}

	fmt.Println("✅ 未使用资源已清理")
	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func resolveDocID(ctx context.Context, client *siyuan.Client, notebookID, path string) (string, error) {
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", err
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
