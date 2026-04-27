package document

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
	"github.com/cicbyte/siyuan-cli/internal/tui"
	"go.uber.org/zap"
)

// CreateMdOptions 定义document createMd命令的选项
type CreateMdOptions struct {
	NotebookIdentifier string // 笔记本ID或名称
	Path               string // 文档路径（不包含文件名），如"/foo/bar"
	Title              string // 文档标题（可选）
	Markdown           string // Markdown内容
	MarkdownFile       string // Markdown文件路径（与Markdown互斥）
}

// CreateMdWithMarkdown 使用Markdown内容创建文档
func CreateMdWithMarkdown(opts CreateMdOptions) error {
	logger := log.GetLogger()
	logger.Info("开始使用Markdown创建文档",
		zap.String("notebook", opts.NotebookIdentifier),
		zap.String("path", opts.Path),
		zap.String("title", opts.Title))

	// 检查思源笔记配置
	if !siyuan.IsSiYuanConfigValid() {
		logger.Error("思源笔记配置无效或未启用")
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli auth login' 配置连接")
		return fmt.Errorf("思源笔记配置无效")
	}
	logger.Info("思源笔记配置验证通过")

	// 验证参数
	if strings.TrimSpace(opts.NotebookIdentifier) == "" {
		err := fmt.Errorf("笔记本标识符不能为空")
		logger.Error("笔记本标识符为空", zap.Error(err))
		fmt.Println("❌ 错误: 笔记本标识符不能为空")
		fmt.Println("💡 使用方法: siyuan-cli document createMd <笔记本名称或ID> [选项]")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
		return err
	}

	// 验证Markdown内容或文件，否则尝试从 stdin 读取
	if strings.TrimSpace(opts.Markdown) == "" && strings.TrimSpace(opts.MarkdownFile) == "" {
		pipeContent, err := output.ReadPipeOrFile("")
		if err != nil {
			logger.Error("读取管道输入失败", zap.Error(err))
			fmt.Printf("❌ 读取输入失败: %v\n", err)
			return err
		}
		if strings.TrimSpace(pipeContent) == "" {
			err := fmt.Errorf("必须提供Markdown内容")
			logger.Error("Markdown内容为空", zap.Error(err))
			fmt.Println("❌ 错误: 必须提供Markdown内容")
			fmt.Println("💡 使用 --content 提供内容、--file 指定文件、或通过管道输入")
			return err
		}
		opts.Markdown = pipeContent
	}

	if strings.TrimSpace(opts.Markdown) != "" && strings.TrimSpace(opts.MarkdownFile) != "" {
		err := fmt.Errorf("不能同时提供Markdown内容和文件路径")
		logger.Error("Markdown内容和文件同时提供", zap.Error(err))
		fmt.Println("❌ 错误: 不能同时提供Markdown内容和文件路径")
		fmt.Println("💡 请使用 --content 或 --file 选项中的一个")
		return err
	}

	// 如果提供了文件路径，读取文件内容
	if strings.TrimSpace(opts.MarkdownFile) != "" {
		content, err := readMarkdownFile(opts.MarkdownFile)
		if err != nil {
			logger.Error("读取Markdown文件失败", zap.String("file", opts.MarkdownFile), zap.Error(err))
			fmt.Printf("❌ 读取文件失败: %v\n", err)
			return err
		}
		opts.Markdown = content
	}

	// 验证输出格式
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

	// 处理路径
	var path string
	if opts.Path == "" {
		logger.Info("路径为空，启动TUI路径选择器")
		selectedPath, err := selectPathWithTUI(targetID, targetName, client, ctx)
		if err != nil {
			logger.Error("路径选择失败", zap.String("error", err.Error()))
			fmt.Printf("❌ 路径选择失败: %v\n", err)
			return fmt.Errorf("路径选择失败: %w", err)
		}
		if selectedPath == "" {
			fmt.Println("❌ 用户取消了路径选择")
			return fmt.Errorf("用户取消了操作")
		}
		path = selectedPath
		logger.Info("用户选择了路径", zap.String("path", path))
	} else {
		path = opts.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}

	// 调用API创建文档
	logger.Info("创建Markdown文档",
		zap.String("notebook_id", targetID),
		zap.String("path", path),
		zap.String("title", opts.Title))

	createOpts := siyuan.CreateDocWithMdOptions{
		Notebook: targetID,
		Path:     path,
		Markdown: opts.Markdown,
		Title:    opts.Title,
	}

	result, err := client.CreateDocWithMd(ctx, createOpts)
	if err != nil {
		logger.Error("创建Markdown文档失败",
			zap.String("error", err.Error()),
			zap.String("notebook_id", targetID),
			zap.String("path", path))
		fmt.Printf("❌ 创建文档失败: %v\n", err)

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本 '%s' (%s) 是否存在\n", targetName, targetID)
			fmt.Printf("   - 请确认文档路径 '%s' 是否有效\n", path)
			fmt.Printf("   - 请确认Markdown内容格式是否正确\n")
		}
		return fmt.Errorf("创建文档失败: %w", err)
	}

	logger.Info("成功创建Markdown文档",
		zap.String("doc_id", result.ID),
		zap.String("doc_name", result.Name),
		zap.String("h_path", result.HPath))

	// 根据输出格式处理结果
	if output.IsJSON("") {
		return outputCreateMdJSON(result, targetName)
	}
	return outputCreateMdTable(result, targetName)
}

// readMarkdownFile 读取Markdown文件内容
func readMarkdownFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取文件内容失败: %w", err)
	}

	return content.String(), nil
}

// outputCreateMdJSON 输出JSON格式
func outputCreateMdJSON(result *siyuan.CreateDocWithMdResponse, notebookName string) error {
	data := struct {
		Notebook string                           `json:"notebook"`
		Document *siyuan.CreateDocWithMdResponse `json:"document"`
	}{
		Notebook: notebookName,
		Document: result,
	}

	output.PrintJSON(data)
	return nil
}

// outputCreateMdTable 输出表格格式
func outputCreateMdTable(result *siyuan.CreateDocWithMdResponse, notebookName string) error {
	fmt.Println("文档创建成功")
	output.PrintTable(
		[]string{"属性", "值"},
		[][]string{
			{"笔记本", notebookName},
			{"文档ID", result.ID},
			{"文档名称", result.Name},
			{"路径", result.HPath},
			{"物理路径", result.Path},
			{"根ID", result.RootID},
		},
	)

	return nil
}

// selectPathWithTUI 使用TUI选择文档路径
func selectPathWithTUI(notebookID, notebookName string, client *siyuan.Client, ctx context.Context) (string, error) {
	logger := log.GetLogger()

	docTree, err := client.ListDocTree(ctx, siyuan.ListDocTreeOptions{
		NotebookID: notebookID,
		Path:       "/",
	})
	if err != nil {
		logger.Error("获取文档树失败", zap.String("error", err.Error()))
		return "", fmt.Errorf("获取文档树失败: %w", err)
	}

	nameMap, err := buildNameMapFromSQL(ctx, client, notebookID)
	if err != nil {
		logger.Warn("SQL查询hpath失败，回退到tree name", zap.Error(err))
		nameMap = buildNameMapFromTree(docTree.Tree)
	}

	pathSelector := ui.NewDocumentPathSelectorTUI(notebookName, notebookID, docTree, nameMap)
	return runPathSelector(pathSelector, logger)
}

// runPathSelector 运行路径选择器
func runPathSelector(pathSelector *ui.DocumentPathSelectorTUI, logger *zap.Logger) (string, error) {
	selectedPath, err := pathSelector.Run()
	if err != nil {
		logger.Error("路径选择TUI运行失败", zap.String("error", err.Error()))
		return "", fmt.Errorf("路径选择TUI运行失败: %w", err)
	}

	if selectedPath == "" {
		logger.Info("用户取消了路径选择")
		return "", nil
	}

	logger.Info("路径选择成功", zap.String("selected_path", selectedPath))
	return selectedPath, nil
}
