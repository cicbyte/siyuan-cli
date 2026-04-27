package notebook

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

// RenameOptions 定义notebook rename命令的选项
type RenameOptions struct {
	CurrentIdentifier string // 当前笔记本名称或ID
	NewName           string // 新笔记本名称
}

// RenameNotebook 执行重命名笔记本的逻辑
func RenameNotebook(opts RenameOptions) error {
	logger := log.GetLogger()
	logger.Info("开始重命名笔记本",
		zap.String("current_identifier", opts.CurrentIdentifier),
		zap.String("new_name", opts.NewName))

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
	if strings.TrimSpace(opts.CurrentIdentifier) == "" {
		err := fmt.Errorf("当前笔记本标识符不能为空")
		logger.Error("当前笔记本标识符为空", zap.Error(err))
		fmt.Println("❌ 错误: 当前笔记本标识符不能为空")
		fmt.Println("💡 使用方法: siyuan-cli notebook rename <当前名称或ID> <新名称>")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
		return err
	}

	if strings.TrimSpace(opts.NewName) == "" {
		err := fmt.Errorf("新笔记本名称不能为空")
		logger.Error("新笔记本名称为空", zap.Error(err))
		fmt.Println("❌ 错误: 新笔记本名称不能为空")
		fmt.Println("💡 使用方法: siyuan-cli notebook rename <当前名称或ID> <新名称>")
		return err
	}

	// 验证新名称长度
	if len(opts.NewName) > 50 {
		err := fmt.Errorf("新笔记本名称过长，最多50个字符")
		logger.Error("新笔记本名称过长", zap.String("new_name", opts.NewName), zap.Int("length", len(opts.NewName)))
		fmt.Printf("❌ 错误: 新笔记本名称过长（%d字符），最多50个字符\n", len(opts.NewName))
		return err
	}

	// 创建客户端
	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		logger.Error("创建思源笔记客户端失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 创建思源笔记客户端失败: %v\n", err)
		return fmt.Errorf("创建客户端失败: %w", err)
	}
	logger.Info("思源笔记客户端客户端创建成功")

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

	// 智能匹配当前笔记本
	targetID, currentName, err := FindNotebook(notebooks, opts.CurrentIdentifier)
	if err != nil {
		logger.Error("当前笔记本匹配失败", zap.String("error", err.Error()), zap.String("identifier", opts.CurrentIdentifier))
		fmt.Printf("❌ %v\n", err)
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看所有可用的笔记本")
		return err
	}

	// 检查新名称是否与现有笔记本重名
	for _, nb := range notebooks {
		if strings.EqualFold(nb.Name, opts.NewName) && nb.ID != targetID {
			logger.Warn("新笔记本名称已存在", zap.String("new_name", opts.NewName), zap.String("existing_id", nb.ID))
			fmt.Printf("⚠️  笔记本名称 '%s' 已被使用 (ID: %s)\n", opts.NewName, nb.ID)
			fmt.Printf("💡 提示: 请选择不同的名称\n")
			return fmt.Errorf("笔记本名称已存在")
		}
	}

	// 检查是否有实际变化
	if currentName == opts.NewName {
		logger.Info("新名称与当前名称相同，无需重命名")
		fmt.Printf("ℹ️  笔记本名称已经是 '%s'，无需重命名\n", currentName)
		return nil
	}

	// 调用重命名API
	logger.Info("开始调用RenameNotebook API",
		zap.String("current_id", targetID),
		zap.String("current_name", currentName),
		zap.String("new_name", opts.NewName))

	err = client.RenameNotebook(ctx, targetID, opts.NewName)
	if err != nil {
		logger.Error("重命名笔记本失败",
			zap.String("error", err.Error()),
			zap.String("current_id", targetID),
			zap.String("current_name", currentName),
			zap.String("new_name", opts.NewName))

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Printf("❌ 重命名笔记本失败: %v\n", err)
			// 提供详细的错误诊断
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本 '%s' (%s) 是否存在\n", currentName, targetID)
			fmt.Printf("   - 请确认新名称 '%s' 是否符合要求\n", opts.NewName)
			fmt.Printf("   - 检查笔记本是否被其他程序占用\n")
		}
		return fmt.Errorf("重命名笔记本失败: %w", err)
	}

	logger.Info("成功重命名笔记本",
		zap.String("current_id", targetID),
		zap.String("old_name", currentName),
		zap.String("new_name", opts.NewName))

	fmt.Println("成功重命名笔记本")
	output.PrintTable(
		[]string{"属性", "值"},
		[][]string{
			{"旧名称", currentName},
			{"新名称", opts.NewName},
			{"ID", targetID},
		},
	)
	fmt.Println("提示: 重命名操作完成后，笔记本的文件系统路径也会相应更新")

	return nil
}