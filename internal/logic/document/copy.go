package document

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

type CopyOptions struct {
	NotebookIdentifier string
	Path               string
}

func CopyDocument(opts CopyOptions) error {
	logger := log.GetLogger()
	logger.Info("复制文档", zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path))

	if strings.TrimSpace(opts.Path) == "" {
		fmt.Println("❌ 错误: 请提供文档路径")
		fmt.Println("💡 使用方法: siyuan-cli document copy <笔记本> <文档路径>")
		return fmt.Errorf("文档路径不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		return err
	}

	targetID, targetName, err := FindNotebook(notebooks, opts.NotebookIdentifier)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return err
	}

	resolvedPath, err := ResolveDocPath(ctx, client, targetID, targetName, opts.Path)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return err
	}

	docID := strings.TrimSuffix(strings.Trim(resolvedPath, "/"), "/")
	parts := strings.Split(docID, "/")
	docID = parts[len(parts)-1]

	err = client.DuplicateDoc(ctx, docID)
	if err != nil {
		fmt.Printf("❌ 复制文档失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已复制文档 '%s'\n", opts.Path)
	return nil
}
