package notebook

import (
	"context"
	"fmt"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

// OpenOptions 定义notebook open命令的选项
type OpenOptions struct {
	NotebookIdentifier string // 笔记本ID或名称
}

// OpenNotebook 执行打开笔记本的逻辑
func OpenNotebook(opts OpenOptions) error {
	logger := log.GetLogger()
	logger.Info("开始打开笔记本", zap.String("identifier", opts.NotebookIdentifier))

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
	if opts.NotebookIdentifier == "" {
		err := fmt.Errorf("笔记本标识符不能为空")
		logger.Error("笔记本标识符为空", zap.Error(err))
		fmt.Println("❌ 错误: 笔记本标识符不能为空")
		fmt.Println("💡 使用方法: siyuan-cli notebook open <笔记本名称或ID>")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
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

	// 检查笔记本是否已经打开
	for _, nb := range notebooks {
		if nb.ID == targetID && !nb.Closed {
			logger.Info("笔记本已经打开", zap.String("notebook_id", targetID), zap.String("notebook_name", targetName))
			fmt.Printf("📂 笔记本 '%s' 已经打开\n", targetName)
			return nil
		}
	}

	// 调用打开笔记本API
	logger.Info("开始调用OpenNotebook API", zap.String("notebook_id", targetID), zap.String("notebook_name", targetName))
	err = client.OpenNotebook(ctx, targetID)
	if err != nil {
		logger.Error("打开笔记本失败", zap.String("error", err.Error()), zap.String("notebook_id", targetID))

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Printf("❌ 打开笔记本失败: %v\n", err)
			// 提供详细的错误诊断
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本 '%s' (%s) 是否存在\n", targetName, targetID)
		}
		return fmt.Errorf("打开笔记本失败: %w", err)
	}

	logger.Info("成功打开笔记本", zap.String("notebook_id", targetID), zap.String("notebook_name", targetName))
	fmt.Printf("✅ 成功打开笔记本: %s (%s)\n", targetName, targetID)
	return nil
}

