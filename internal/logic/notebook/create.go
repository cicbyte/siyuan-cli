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

// CreateOptions 定义notebook create命令的选项
type CreateOptions struct {
	Name string // 笔记本名称（必需）
	Icon string // 笔记本图标（可选）
}

// CreateNotebook 执行创建笔记本的逻辑
func CreateNotebook(opts CreateOptions) error {
	logger := log.GetLogger()
	logger.Info("开始创建笔记本", zap.String("name", opts.Name), zap.String("icon", opts.Icon))

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
	if strings.TrimSpace(opts.Name) == "" {
		err := fmt.Errorf("笔记本名称不能为空")
		logger.Error("笔记本名称为空", zap.Error(err))
		fmt.Println("❌ 错误: 笔记本名称不能为空")
		fmt.Println("💡 使用方法: siyuan-cli notebook create <笔记本名称>")
		return err
	}

	// 验证笔记本名称长度
	if len(opts.Name) > 50 {
		err := fmt.Errorf("笔记本名称过长，最多50个字符")
		logger.Error("笔记本名称过长", zap.String("name", opts.Name), zap.Int("length", len(opts.Name)))
		fmt.Printf("❌ 错误: 笔记本名称过长（%d字符），最多50个字符\n", len(opts.Name))
		return err
	}

	// 验证图标（如果提供）
	if opts.Icon != "" && len(opts.Icon) > 10 {
		err := fmt.Errorf("图标过长，最多10个字符")
		logger.Error("图标过长", zap.String("icon", opts.Icon), zap.Int("length", len(opts.Icon)))
		fmt.Printf("❌ 错误: 图标过长（%d字符），最多10个字符\n", len(opts.Icon))
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

	// 检查笔记本名称是否已存在
	logger.Info("检查笔记本名称是否已存在")
	existingNotebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		logger.Error("获取笔记本列表失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		return fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	for _, nb := range existingNotebooks {
		if strings.EqualFold(nb.Name, opts.Name) {
			logger.Warn("笔记本名称已存在", zap.String("name", opts.Name), zap.String("existing_id", nb.ID))
			fmt.Printf("⚠️  笔记本 '%s' 已存在 (ID: %s)\n", nb.Name, nb.ID)
			if nb.Closed {
				fmt.Printf("💡 提示: 该笔记本已关闭，可以使用 'siyuan-cli notebook open %s' 重新打开\n", opts.Name)
			} else {
				fmt.Printf("💡 提示: 该笔记本当前是打开状态\n")
			}
			return fmt.Errorf("笔记本名称已存在")
		}
	}

	// 调用创建笔记本API
	logger.Info("开始调用CreateNotebook API")
	var newNotebook *siyuan.Notebook

	if opts.Icon != "" {
		// 尝试创建带图标的笔记本
		logger.Info("尝试创建带图标的笔记本", zap.String("icon", opts.Icon))
		newNotebook, err = client.CreateNotebookWithIcon(ctx, opts.Name, opts.Icon)

		if err != nil {
			logger.Warn("创建带图标的笔记本失败，回退到普通创建", zap.String("error", err.Error()))
			fmt.Printf("⚠️  创建带图标的笔记本失败，将创建普通笔记本\n")
			fmt.Printf("   错误信息: %v\n", err)

			// 回退到普通创建
			newNotebook, err = client.CreateNotebook(ctx, opts.Name)
		}
	} else {
		// 创建普通笔记本
		newNotebook, err = client.CreateNotebook(ctx, opts.Name)
	}

	if err != nil {
		logger.Error("创建笔记本失败", zap.String("error", err.Error()), zap.String("name", opts.Name))

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Printf("❌ 创建笔记本失败: %v\n", err)
			// 提供详细的错误诊断
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本名称 '%s' 是否符合要求\n", opts.Name)
			fmt.Printf("   - 笔记本名称不能包含特殊字符\n")
		}
		return fmt.Errorf("创建笔记本失败: %w", err)
	}

	logger.Info("成功创建笔记本", zap.String("name", newNotebook.Name), zap.String("id", newNotebook.ID))

	fmt.Println("成功创建笔记本")
	output.PrintTable(
		[]string{"属性", "值"},
		[][]string{
			{"名称", newNotebook.Name},
			{"ID", newNotebook.ID},
			{"图标", newNotebook.Icon},
		},
	)
	fmt.Printf("💡 提示: 使用 'siyuan-cli notebook open %s' 打开新笔记本\n", newNotebook.Name)

	return nil
}