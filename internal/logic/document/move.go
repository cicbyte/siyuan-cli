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

type MoveOptions struct {
	NotebookIdentifier string
	SrcPath            string
	DestPath           string
}

func MoveDocument(opts MoveOptions) error {
	logger := log.GetLogger()
	logger.Info("移动文档",
		zap.String("notebook", opts.NotebookIdentifier),
		zap.String("src", opts.SrcPath),
		zap.String("dest", opts.DestPath))

	for _, p := range []string{opts.SrcPath, opts.DestPath} {
		if strings.TrimSpace(p) == "" {
			fmt.Println("❌ 错误: 源路径和目标路径都不能为空")
			fmt.Println("💡 使用方法: siyuan-cli document move <笔记本> <源路径> <目标路径>")
			return fmt.Errorf("路径不能为空")
		}
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

	srcResolved, err := ResolveDocPath(ctx, client, targetID, targetName, opts.SrcPath)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return err
	}

	// 目标路径支持: notebook2:/路径 或 notebook2（移到目标笔记本根目录）
	var destNotebookID, destNotebookName, destPath string
	destPath = opts.DestPath
	if parts := strings.SplitN(opts.DestPath, ":", 2); len(parts) == 2 {
		nbID, nbName, nbErr := FindNotebook(notebooks, parts[0])
		if nbErr != nil {
			fmt.Printf("❌ 目标笔记本未找到: %v\n", nbErr)
			return nbErr
		}
		destNotebookID = nbID
		destNotebookName = nbName
		destPath = parts[1]
	} else if nbID, nbName, nbErr := FindNotebook(notebooks, opts.DestPath); nbErr == nil {
		destNotebookID = nbID
		destNotebookName = nbName
		destPath = "/"
	} else {
		destNotebookID = targetID
		destNotebookName = targetName
	}

	destResolved, err := ResolveDocPath(ctx, client, destNotebookID, destNotebookName, destPath)
	if err != nil {
		fmt.Printf("❌ 目标路径解析失败: %v\n", err)
		return err
	}

	err = client.MoveDocs(ctx, siyuan.MoveDocsOptions{
		FromPaths: []string{srcResolved},
		ToNotebook: destNotebookID,
		ToPath:    destResolved,
	})
	if err != nil {
		fmt.Printf("❌ 移动文档失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已将 '%s' 移动到 '%s'（笔记本: %s）\n", opts.SrcPath, opts.DestPath, targetName)
	return nil
}
