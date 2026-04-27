package fav

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/cicbyte/siyuan-cli/internal/utils"
	"go.uber.org/zap"
)

// FavOptions 定义fav命令的选项
type FavOptions struct {
	Content string // 收藏的内容（第一个参数）
}

// AddToFavorites 将内容添加到收藏笔记本
func AddToFavorites(opts FavOptions) error {
	logger := log.GetLogger()
	logger.Info("开始添加收藏")

	// 检查思源笔记配置
	if !siyuan.IsSiYuanConfigValid() {
		logger.Error("思源笔记配置无效或未启用")
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli siyuan config' 查看配置")
		fmt.Println("请运行 'siyuan-cli siyuan set enabled true' 启用功能")
		return fmt.Errorf("思源笔记配置无效")
	}
	logger.Info("思源笔记配置验证通过")

	// 获取内容（优先使用管道输入，其次使用参数）
	content := getContent(opts.Content)
	if strings.TrimSpace(content) == "" {
		err := fmt.Errorf("没有提供内容")
		logger.Error("内容为空", zap.Error(err))
		fmt.Println("❌ 错误: 没有提供内容")
		fmt.Println("💡 使用方法:")
		fmt.Println("   siyuan-cli fav \"要收藏的内容\"")
		fmt.Println("   echo \"要收藏的内容\" | siyuan-cli fav")
		return err
	}

	logger.Info("获取到内容", zap.String("content_length", fmt.Sprintf("%d", len(content))))

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

	// 确保收藏笔记本存在
	notebookID, err := ensureFavNotebook(ctx, client)
	if err != nil {
		logger.Error("确保收藏笔记本失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 确保收藏笔记本失败: %v\n", err)
		return err
	}

	logger.Info("收藏笔记本准备完成", zap.String("notebook_id", notebookID))

	// 生成路径和标题
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	path := fmt.Sprintf("/%s/%s", year, month)
	title := extractTitle(content, now)

	logger.Info("生成路径和标题",
		zap.String("year", year),
		zap.String("month", month),
		zap.String("path", path),
		zap.String("title", title))

	// 创建文档
	createOpts := siyuan.CreateDocWithMdOptions{
		Notebook:  notebookID,
		Path:      path,
		Markdown:  content,
		Title:     title,
	}

	result, err := client.CreateDocWithMd(ctx, createOpts)
	if err != nil {
		logger.Error("创建收藏文档失败",
			zap.String("error", err.Error()),
			zap.String("path", path),
			zap.String("title", title))
		fmt.Printf("❌ 创建收藏文档失败: %v\n", err)

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认收藏笔记本是否可访问\n")
			fmt.Printf("   - 请确认内容格式是否正确\n")
		}
		return fmt.Errorf("创建收藏文档失败: %w", err)
	}

	logger.Info("成功创建收藏文档",
		zap.String("doc_id", result.ID),
		zap.String("doc_name", result.Name),
		zap.String("doc_path", result.HPath))

	// 显示成功信息
	fmt.Println("收藏成功")
	output.PrintTable(
		[]string{"属性", "值"},
		[][]string{
			{"笔记本", "我的收藏"},
			{"文档", result.Name},
			{"路径", result.HPath},
			{"时间", now.Format("2006-01-02 15:04:05")},
		},
	)

	return nil
}

// getContent 获取内容，优先从管道读取，其次使用参数
func getContent(paramContent string) string {
	// 检查是否有管道输入
	if utils.IsInputFromPipe() {
		// 从管道读取内容
		content, err := readFromPipe()
		if err != nil {
			fmt.Printf("❌ 读取管道内容失败: %v\n", err)
			return ""
		}
		return content
	}

	// 使用参数内容
	return paramContent
}

// readFromPipe 从管道读取内容
func readFromPipe() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// 有管道输入
		reader := bufio.NewReader(os.Stdin)
		var buffer bytes.Buffer

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", err
			}
			buffer.WriteString(line)
		}

		return buffer.String(), nil
	}

	return "", nil
}

// ensureFavNotebook 确保收藏笔记本存在，如果不存在则创建
func ensureFavNotebook(ctx context.Context, client *siyuan.Client) (string, error) {
	logger := log.GetLogger()
	notebookName := "我的收藏"

	// 获取所有笔记本
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	// 查找收藏笔记本
	for _, notebook := range notebooks {
		if notebook.Name == notebookName {
			return notebook.ID, nil
		}
	}

	// 如果没有找到，自动创建新的笔记本
	fmt.Printf("📝 未找到 '%s' 笔记本，正在自动创建...\n", notebookName)

	// 调用创建笔记本API，采用和notebook create相同的逻辑
	logger.Info("开始调用CreateNotebook API创建收藏笔记本")
	var newNotebook *siyuan.Notebook

	// 尝试创建带图标的笔记本
	logger.Info("尝试创建带图标的笔记本", zap.String("icon", "⭐"))
	newNotebook, err = client.CreateNotebookWithIcon(ctx, notebookName, "⭐")

	if err != nil {
		logger.Warn("创建带图标的笔记本失败，回退到普通创建", zap.String("error", err.Error()))
		fmt.Printf("⚠️  创建带图标的笔记本失败，将创建普通笔记本\n")
		fmt.Printf("   错误信息: %v\n", err)

		// 回退到普通创建
		newNotebook, err = client.CreateNotebook(ctx, notebookName)
	}

	if err != nil {
		logger.Error("创建收藏笔记本失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 创建 '%s' 笔记本失败: %v\n", notebookName, err)

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Printf("💡 请手动在思源笔记中创建名为 '%s' 的笔记本\n", notebookName)
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本名称 '%s' 是否符合要求\n", notebookName)
			fmt.Printf("   - 笔记本名称不能包含特殊字符\n")
		}
		return "", fmt.Errorf("创建收藏笔记本失败: %w", err)
	}

	fmt.Printf("✅ 成功创建 '%s' 笔记本 (ID: %s)\n", notebookName, newNotebook.ID)
	if newNotebook.Icon != "" {
		fmt.Printf("🎨 图标: %s\n", newNotebook.Icon)
	}
	return newNotebook.ID, nil
}

// extractTitle 从内容中提取标题，如果没有找到则使用时间戳
func extractTitle(content string, now time.Time) string {
	// 尝试提取第一个 # 标题
	lines := strings.Split(content, "\n")
	titlePattern := regexp.MustCompile(`^#\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := titlePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			if title != "" {
				// 清理标题，移除无效字符
				title = strings.ReplaceAll(title, "/", "-")
				title = strings.ReplaceAll(title, "\\", "-")
				title = strings.ReplaceAll(title, ":", "-")
				if len(title) > 50 {
					title = title[:50] + "..."
				}
				return title
			}
		}
	}

	// 如果没有找到标题，使用时间戳
	return now.Format("2006-01-02 15-04-05")
}