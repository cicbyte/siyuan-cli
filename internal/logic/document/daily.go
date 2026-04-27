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

type DailyOptions struct {
	NotebookIdentifier string
}

func CreateDailyNote(opts DailyOptions) error {
	logger := log.GetLogger()
	logger.Info("创建日记", zap.String("notebook", opts.NotebookIdentifier))

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if opts.NotebookIdentifier == "" {
		// 查找第一个笔记本作为默认
		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
			return err
		}
		for _, nb := range notebooks {
			if !nb.Closed {
				opts.NotebookIdentifier = nb.ID
				break
			}
		}
		if opts.NotebookIdentifier == "" {
			fmt.Println("❌ 错误: 没有打开的笔记本，请使用 --notebook 指定")
			return fmt.Errorf("没有打开的笔记本")
		}
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

	result, err := client.CreateDailyNote(ctx, siyuan.CreateDailyNoteOptions{Notebook: targetID})
	if err != nil {
		fmt.Printf("❌ 创建日记失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 日记已创建\n")
	fmt.Printf("   笔记本: %s\n", targetName)
	fmt.Printf("   标题: %s\n", result.Name)
	fmt.Printf("   路径: %s\n", strings.TrimPrefix(result.HPath, "/"))
	return nil
}
