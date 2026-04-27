package importcmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

type ImportMdOptions struct {
	File               string
	NotebookIdentifier string
	DestPath           string
}

func ImportMd(opts ImportMdOptions) error {
	logger := log.GetLogger()
	logger.Info("导入 Markdown", zap.String("file", opts.File), zap.String("notebook", opts.NotebookIdentifier))

	if strings.TrimSpace(opts.File) == "" {
		fmt.Println("❌ 错误: 请提供文件或目录路径")
		fmt.Println("💡 使用方法: siyuan-cli import md <文件或目录> --notebook <笔记本>")
		return fmt.Errorf("文件路径不能为空")
	}

	if _, err := os.Stat(opts.File); os.IsNotExist(err) {
		fmt.Printf("❌ 文件或目录不存在: %s\n", opts.File)
		return err
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

	importOpts := siyuan.ImportStdMdOptions{
		Notebook: targetID,
		Path:     opts.DestPath,
	}

	// 检查是否是 zip 文件
	if strings.HasSuffix(strings.ToLower(opts.File), ".zip") {
		if err := client.ImportZipMd(ctx, siyuan.ImportZipMdOptions{
			Notebook: targetID,
			Path:     opts.DestPath,
		}); err != nil {
			fmt.Printf("❌ 导入 ZIP Markdown 失败: %v\n", err)
			return err
		}
	} else {
		if err := client.ImportStdMd(ctx, importOpts); err != nil {
			fmt.Printf("❌ 导入 Markdown 失败: %v\n", err)
			return err
		}
	}

	fmt.Printf("✅ 已导入到笔记本 '%s'\n", targetName)
	return nil
}

type ImportSyOptions struct {
	File               string
	NotebookIdentifier string
}

func ImportSy(opts ImportSyOptions) error {
	logger := log.GetLogger()
	logger.Info("导入思源格式", zap.String("file", opts.File), zap.String("notebook", opts.NotebookIdentifier))

	if strings.TrimSpace(opts.File) == "" {
		fmt.Println("❌ 错误: 请提供文件或目录路径")
		fmt.Println("💡 使用方法: siyuan-cli import sy <文件或目录> --notebook <笔记本>")
		return fmt.Errorf("文件路径不能为空")
	}

	if _, err := os.Stat(opts.File); os.IsNotExist(err) {
		fmt.Printf("❌ 文件或目录不存在: %s\n", opts.File)
		return err
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

	if err := client.ImportSY(ctx, siyuan.ImportSYOptions{
		Notebook: targetID,
	}); err != nil {
		fmt.Printf("❌ 导入思源格式失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已导入到笔记本 '%s'\n", targetName)
	return nil
}
