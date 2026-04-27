package block

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

type UpdateOptions struct {
	ID      string
	Content string
}

func UpdateBlock(opts UpdateOptions) error {
	logger := log.GetLogger()
	logger.Info("更新块", zap.String("id", opts.ID))

	if strings.TrimSpace(opts.ID) == "" {
		fmt.Println("❌ 错误: 请提供块 ID")
		fmt.Println("💡 使用方法: siyuan-cli block update <block-id> --content \"新内容\"")
		return fmt.Errorf("块 ID 不能为空")
	}

	if strings.TrimSpace(opts.Content) == "" {
		fmt.Println("❌ 错误: 请提供新内容")
		fmt.Println("💡 使用 --content 指定新内容，或通过管道传入")
		return fmt.Errorf("内容不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.UpdateBlock(ctx, siyuan.UpdateBlockOptions{
		ID:       opts.ID,
		DataType: "markdown",
		Data:     opts.Content,
	})
	if err != nil {
		fmt.Printf("❌ 更新块失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 块 %s 已更新\n", opts.ID)
	return nil
}

type AppendOptions struct {
	DocID    string // 直接传文档 ID
	NotebookIdentifier string
	Path               string
	Content            string
	DataType           string
}

func AppendBlock(opts AppendOptions) error {
	logger := log.GetLogger()
	logger.Info("追加块", zap.String("docID", opts.DocID), zap.String("notebook", opts.NotebookIdentifier), zap.String("path", opts.Path))

	if strings.TrimSpace(opts.Content) == "" {
		fmt.Println("❌ 错误: 请提供内容")
		fmt.Println("💡 使用 --content 指定内容，或通过管道传入")
		return fmt.Errorf("内容不能为空")
	}

	dataType := opts.DataType
	if dataType == "" {
		dataType = "markdown"
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var docID string
	// 直接传文档 ID
	if opts.DocID != "" {
		docID = opts.DocID
	} else if opts.NotebookIdentifier != "" && opts.Path != "" {
		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
			return err
		}

		targetID, _, err := findNotebook(notebooks, opts.NotebookIdentifier)
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
	} else {
		fmt.Println("❌ 错误: 请提供文档 ID 或 笔记本/文档路径")
		fmt.Println("💡 使用方法: siyuan-cli block append <doc-id> --content \"内容\"")
		fmt.Println("   或: siyuan-cli block append <笔记本/文档路径> --content \"内容\"")
		return fmt.Errorf("请提供文档 ID 或 笔记本/文档路径")
	}

	blockID, err := client.AppendBlock(ctx, siyuan.AppendBlockOptions{
		ParentID: docID,
		DataType: dataType,
		Data:     opts.Content,
	})
	if err != nil {
		fmt.Printf("❌ 追加块失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已追加内容（新块 ID: %s）\n", blockID)
	return nil
}

type DeleteOptions struct {
	ID    string
	Force bool
}

func DeleteBlock(opts DeleteOptions) error {
	logger := log.GetLogger()
	logger.Info("删除块", zap.String("id", opts.ID))

	if strings.TrimSpace(opts.ID) == "" {
		fmt.Println("❌ 错误: 请提供块 ID")
		fmt.Println("💡 使用方法: siyuan-cli block delete <block-id>")
		return fmt.Errorf("块 ID 不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.DeleteBlock(ctx, opts.ID)
	if err != nil {
		fmt.Printf("❌ 删除块失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 块 %s 已删除\n", opts.ID)
	return nil
}
