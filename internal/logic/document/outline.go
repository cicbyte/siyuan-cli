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

type OutlineOptions struct {
	DocID              string
	NotebookIdentifier string
	Path               string
}

func GetDocumentOutline(opts OutlineOptions) error {
	logger := log.GetLogger()
	logger.Info("获取文档大纲", zap.String("docID", opts.DocID), zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path))

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
		if strings.TrimSpace(opts.Path) == "" {
			fmt.Println("❌ 错误: 请提供文档路径")
			fmt.Println("💡 使用方法: siyuan-cli document outline <笔记本> <文档路径>")
			return fmt.Errorf("文档路径不能为空")
		}

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

		docID = strings.TrimSuffix(strings.Trim(resolvedPath, "/"), "/")
		parts := strings.Split(docID, "/")
		docID = parts[len(parts)-1]
	}

	outline, err := client.GetDocOutline(ctx, docID)
	if err != nil {
		fmt.Printf("❌ 获取大纲失败: %v\n", err)
		return err
	}

	if len(outline) == 0 {
		fmt.Println("该文档没有大纲（无标题）")
		return nil
	}

	fmt.Println("文档大纲:")
	for _, item := range outline {
		indent := strings.Repeat("  ", item.Depth)
		content := strings.ReplaceAll(item.Content, "\n", " ")
		fmt.Printf("%s%s %s\n", indent, strings.Repeat("#", item.Level), content)
	}
	return nil
}
