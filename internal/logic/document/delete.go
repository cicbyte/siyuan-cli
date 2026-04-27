package document

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

// DeleteOptions 定义document delete命令的选项
type DeleteOptions struct {
	NotebookIdentifier string // 笔记本ID或名称
	DocumentPath       string // 文档路径（如"/foo/bar/doc"）
}

// DeleteDocument 删除文档
func DeleteDocument(opts DeleteOptions) error {
	logger := log.GetLogger()
	logger.Info("开始删除文档",
		zap.String("notebook", opts.NotebookIdentifier),
		zap.String("document_path", opts.DocumentPath))

	// 检查思源笔记配置
	if !siyuan.IsSiYuanConfigValid() {
		logger.Error("思源笔记配置无效或未启用")
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli siyuan config' 查看配置")
		fmt.Println("请运行 'siyuan-cli siyuan set enabled true' 启用功能")
		return fmt.Errorf("思源笔记配置无效")
	}
	logger.Info("思源笔记配置验证通过")

	// 验证参数
	if strings.TrimSpace(opts.NotebookIdentifier) == "" {
		err := fmt.Errorf("笔记本标识符不能为空")
		logger.Error("笔记本标识符为空", zap.Error(err))
		fmt.Println("❌ 错误: 笔记本标识符不能为空")
		fmt.Println("💡 使用方法: siyuan-cli document delete <笔记本名称或ID> <文档路径>")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
		return err
	}

	if strings.TrimSpace(opts.DocumentPath) == "" {
		err := fmt.Errorf("文档路径不能为空")
		logger.Error("文档路径为空", zap.Error(err))
		fmt.Println("❌ 错误: 文档路径不能为空")
		fmt.Println("💡 使用方法: siyuan-cli document delete <笔记本名称或ID> <文档路径>")
		fmt.Println("💡 文档路径示例: /项目文档/API文档")
		return err
	}

	// 创建客户端
	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		logger.Error("创建思源笔记客户端失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 创建思源笔记客户端失败: %v\n", err)
		return fmt.Errorf("创建客户端失败: %w", err)
	}
	logger.Info("思源笔记客户端创建成功")

	// 创建带超时的Context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取所有笔记本列表用于匹配
	logger.Info("获取笔记本列表进行匹配")
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		logger.Error("获取笔记本列表失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		return fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	// 智能匹配笔记本
	targetID, targetName, err := FindNotebook(notebooks, opts.NotebookIdentifier)
	if err != nil {
		logger.Error("笔记本匹配失败", zap.String("error", err.Error()), zap.String("identifier", opts.NotebookIdentifier))
		fmt.Printf("❌ %v\n", err)
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看所有可用的笔记本")
		return err
	}

	logger.Info("成功匹配笔记本",
		zap.String("id", targetID),
		zap.String("name", targetName))

	// 规范化文档路径
	originalPath := opts.DocumentPath
	logger.Info("原始接收到的路径", zap.String("original_path", originalPath))

	documentPath := originalPath
	if !strings.HasPrefix(documentPath, "/") {
		documentPath = "/" + documentPath
	}

	logger.Info("规范化后的路径", zap.String("normalized_path", documentPath))

	// 使用getIDsByHPath API根据人类可读路径获取文档ID
	logger.Info("根据人类可读路径获取文档ID",
		zap.String("notebook_id", targetID),
		zap.String("human_path", documentPath))

	docIDs, err := client.GetIDsByHPath(ctx, siyuan.GetIDsByHPathOptions{
		Notebook: targetID,
		Path:     documentPath,
	})
	if err != nil {
		logger.Error("获取文档ID失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 获取文档ID失败: %v\n", err)
		return fmt.Errorf("获取文档ID失败: %w", err)
	}

	if len(docIDs) == 0 {
		err := fmt.Errorf("未找到文档路径 '%s' 对应的文档", documentPath)
		logger.Error("未找到文档", zap.String("human_path", documentPath), zap.Error(err))
		fmt.Printf("❌ 错误: %v\n", err)
		fmt.Println("💡 请确认:")
		fmt.Printf("   - 文档路径 '%s' 是否存在\n", documentPath)
		fmt.Printf("   - 可以使用 'siyuan-cli document ls %s' 查看可用文档\n", targetName)
		return err
	}

	if len(docIDs) > 1 {
		logger.Warn("找到多个匹配的文档，使用第一个",
			zap.String("human_path", documentPath),
			zap.Strings("doc_ids", docIDs))
	}

	docID := docIDs[0]
	logger.Info("找到文档ID", zap.String("doc_id", docID))

	// 现在我们有了文档ID，需要获取其物理路径
	// 删除API需要的是相对于笔记本的路径，格式为 /20231201120000-abc123.sy
	deletePath := "/" + docID + ".sy"
	logger.Info("构建删除路径", zap.String("delete_path", deletePath))

	logger.Info("删除文档",
		zap.String("notebook_id", targetID),
		zap.String("doc_id", docID),
		zap.String("delete_path", deletePath))

	deleteOpts := siyuan.RemoveDocOptions{
		NotebookID: targetID,
		Path:       deletePath,
	}

	err = client.RemoveDoc(ctx, deleteOpts)
	if err != nil {
		logger.Error("删除文档失败",
			zap.String("error", err.Error()),
			zap.String("notebook_id", targetID),
			zap.String("document_path", documentPath))
		fmt.Printf("❌ 删除文档失败: %v\n", err)

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本 '%s' (%s) 是否存在\n", targetName, targetID)
			fmt.Printf("   - 请确认文档路径 '%s' 是否存在且为文档文件\n", documentPath)
		}
		return fmt.Errorf("删除文档失败: %w", err)
	}

	logger.Info("成功删除文档",
		zap.String("doc_id", docID),
		zap.String("notebook", targetName),
		zap.String("document_path", documentPath))

	// 显示简单的成功信息
	fmt.Println("文档删除成功")
	output.PrintTable(
		[]string{"属性", "值"},
		[][]string{
			{"笔记本", targetName},
			{"路径", documentPath},
		},
	)

	return nil
}