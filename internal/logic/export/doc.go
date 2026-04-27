package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/cicbyte/siyuan-cli/internal/ui"
	"go.uber.org/zap"
)

type ExportDocOptions struct {
	DocID              string
	NotebookIdentifier string
	Path               string
	Format             string
	OutputFile         string
}

func ExportDoc(opts ExportDocOptions) error {
	logger := log.GetLogger()
	logger.Info("导出文档", zap.String("docID", opts.DocID), zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path), zap.String("format", opts.Format))

	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "md"
	}
	validFormats := map[string]bool{"md": true, "html": true, "docx": true}
	if !validFormats[format] {
		fmt.Printf("❌ 不支持的格式: %s（支持: md, html, docx）\n", opts.Format)
		return fmt.Errorf("不支持的格式: %s", opts.Format)
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var docID string
	if opts.DocID != "" {
		docID = opts.DocID
	} else {
		if strings.TrimSpace(opts.Path) == "" {
			fmt.Println("❌ 错误: 请提供文档路径")
			fmt.Println("💡 使用方法: siyuan-cli export doc <笔记本> <文档路径> --format <格式>")
			return fmt.Errorf("文档路径不能为空")
		}

		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
			return err
		}

		targetID, _, err := document.FindNotebook(notebooks, opts.NotebookIdentifier)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return err
		}

		var resolveErr error
		docID, resolveErr = resolveDocID(ctx, client, targetID, opts.Path)
		if resolveErr != nil {
			fmt.Printf("❌ %v\n", resolveErr)
			return resolveErr
		}
	}

	switch format {
	case "md":
		content, _, err := client.ExportMdContent(ctx, docID)
		if err != nil {
			fmt.Printf("❌ 导出 Markdown 失败: %v\n", err)
			return err
		}
		outputPath := opts.OutputFile
		if outputPath == "" {
			outputPath = docID + ".md"
		}
		if err := writeToFile(outputPath, []byte(content)); err != nil {
			return err
		}
		fmt.Printf("✅ 已导出 Markdown: %s\n", outputPath)

	case "html":
		html, err := client.ExportHTML(ctx, siyuan.ExportDocOptions{ID: docID})
		if err != nil {
			fmt.Printf("❌ 导出 HTML 失败: %v\n", err)
			return err
		}
		outputPath := opts.OutputFile
		if outputPath == "" {
			outputPath = docID + ".html"
		}
		if err := writeToFile(outputPath, []byte(html)); err != nil {
			return err
		}
		fmt.Printf("✅ 已导出 HTML: %s\n", outputPath)

	case "docx":
		_, err := client.ExportDocx(ctx, siyuan.ExportDocOptions{ID: docID})
		if err != nil {
			fmt.Printf("❌ 导出 Word 失败: %v\n", err)
			return err
		}
		fmt.Printf("✅ 已导出 Word 文档（请到思源导出目录查看）\n")
	}

	return nil
}

type ExportNotebookOptions struct {
	NotebookIdentifier string
	Format             string
	OutputDir          string
}

func ExportNotebook(opts ExportNotebookOptions) error {
	logger := log.GetLogger()
	logger.Info("导出笔记本", zap.String("notebook", opts.NotebookIdentifier), zap.String("format", opts.Format), zap.String("output_dir", opts.OutputDir))

	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "md"
	}
	if format != "md" && format != "sy" {
		fmt.Printf("❌ 不支持的格式: %s（支持: md, sy）\n", opts.Format)
		return fmt.Errorf("不支持的格式: %s", opts.Format)
	}

	if strings.TrimSpace(opts.OutputDir) == "" {
		fmt.Println("❌ 错误: 请指定输出目录 (-o)")
		fmt.Println("💡 使用方法: siyuan-cli export notebook <笔记本> -o <输出目录>")
		return fmt.Errorf("输出目录不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		return err
	}

	targetID, targetName, err := document.FindNotebook(notebooks, opts.NotebookIdentifier)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return err
	}

	// 阶段1: 调用导出 API（带 spinner）
	var exportResult *siyuan.ExportNotebookMdResponse
	err = ui.RunSpinner(fmt.Sprintf("正在导出笔记本 '%s' ...", targetName), func() error {
		if format == "md" {
			exportResult, err = client.ExportNotebookMd(ctx, siyuan.ExportNotebookMdOptions{Notebook: targetID})
		} else {
			exportResult, err = client.ExportNotebookSY(ctx, siyuan.ExportNotebookSYOptions{ID: targetID})
		}
		return err
	})
	if err != nil {
		fmt.Printf("❌ 导出失败: %v\n", err)
		return err
	}
	if exportResult == nil || exportResult.Zip == "" {
		fmt.Println("❌ 导出失败: 未获取到导出文件信息")
		return fmt.Errorf("导出响应为空")
	}

	// 阶段2: 下载 zip（带进度条）
	outputPath := filepath.Join(opts.OutputDir, exportResult.Name)
	var zipData []byte
	err = ui.PrintDownloadProgress(fmt.Sprintf("正在下载 %s", exportResult.Name), func(onProgress ui.ProgressFunc) error {
		var downloadErr error
		zipData, downloadErr = client.DownloadExport(ctx, exportResult.Name, onProgress)
		return downloadErr
	})
	if err != nil {
		fmt.Printf("❌ 下载失败: %v\n", err)
		return err
	}

	// 保存文件
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		fmt.Printf("❌ 创建目录失败: %v\n", err)
		return err
	}
	if err := os.WriteFile(outputPath, zipData, 0644); err != nil {
		fmt.Printf("❌ 写入文件失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已导出笔记本 '%s'（格式: %s）→ %s\n", targetName, format, outputPath)
	return nil
}

func writeToFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}
