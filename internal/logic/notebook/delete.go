package notebook

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

// DeleteOptions 定义notebook delete命令的选项
type DeleteOptions struct {
	NotebookIdentifier string // 笔记本ID或名称
	Force              bool   // 是否强制删除，不询问确认
}

// DeleteNotebook 执行删除笔记本的逻辑
func DeleteNotebook(opts DeleteOptions) error {
	logger := log.GetLogger()
	logger.Info("开始删除笔记本", zap.String("identifier", opts.NotebookIdentifier), zap.Bool("force", opts.Force))

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
		fmt.Println("💡 使用方法: siyuan-cli notebook delete <笔记本名称或ID>")
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

	// 获取目标笔记本详细信息
	var targetNotebook *siyuan.Notebook
	for _, nb := range notebooks {
		if nb.ID == targetID {
			targetNotebook = &nb
			break
		}
	}

	// 安全警告和确认
	if !opts.Force {
		fmt.Printf("⚠️  警告：您即将删除笔记本 '%s' (ID: %s)\n", targetName, targetID)
		fmt.Printf("📋 此操作不可逆，笔记本中的所有文档和数据将被永久删除！\n")

		if targetNotebook != nil && !targetNotebook.Closed {
			fmt.Printf("📂 注意：该笔记本当前处于打开状态\n")
		}

		fmt.Printf("\n❓ 确认要删除这个笔记本吗？(输入 'yes' 确认): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "yes" {
			logger.Info("用户取消删除操作", zap.String("notebook", targetName))
			fmt.Println("❌ 删除操作已取消")
			return nil
		}
	} else {
		fmt.Printf("⚠️  强制删除模式：正在删除笔记本 '%s' (ID: %s)\n", targetName, targetID)
	}

	// 调用删除笔记本API
	logger.Info("开始调用RemoveNotebook API",
		zap.String("notebook_id", targetID),
		zap.String("notebook_name", targetName))

	err = client.RemoveNotebook(ctx, targetID)
	if err != nil {
		logger.Error("删除笔记本失败",
			zap.String("error", err.Error()),
			zap.String("notebook_id", targetID),
			zap.String("notebook_name", targetName))

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Printf("❌ 删除笔记本失败: %v\n", err)
			// 提供详细的错误诊断
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本 '%s' (%s) 是否存在\n", targetName, targetID)
			fmt.Printf("   - 请确认您是否有删除该笔记本的权限\n")
			fmt.Printf("   - 检查笔记本是否被其他程序占用\n")
			fmt.Printf("   - 确保笔记本中没有正在进行的同步操作\n")
		}
		return fmt.Errorf("删除笔记本失败: %w", err)
	}

	logger.Info("成功删除笔记本",
		zap.String("notebook_id", targetID),
		zap.String("notebook_name", targetName))

	fmt.Println("成功删除笔记本")
	output.PrintTable(
		[]string{"属性", "值"},
		[][]string{
			{"名称", targetName},
			{"ID", targetID},
		},
	)
	fmt.Println("笔记本及其所有内容已从思源笔记中永久删除")

	return nil
}