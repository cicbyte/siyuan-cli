package notebook

import (
	"context"
	"encoding/json"
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

// GetConfOptions 定义notebook getconf命令的选项
type GetConfOptions struct {
	NotebookIdentifier string // 笔记本ID或名称
	OutputFile         string // 输出文件路径
}

// GetNotebookConf 执行获取笔记本配置的逻辑
func GetNotebookConf(opts GetConfOptions) error {
	logger := log.GetLogger()
	logger.Info("开始获取笔记本配置",
		zap.String("identifier", opts.NotebookIdentifier),
		zap.String("output_file", opts.OutputFile))

	if !siyuan.IsSiYuanConfigValid() {
		logger.Error("思源笔记配置无效或未启用")
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli auth login' 配置连接")
		return fmt.Errorf("思源笔记配置无效")
	}
	logger.Info("思源笔记配置验证通过")

	if strings.TrimSpace(opts.NotebookIdentifier) == "" {
		err := fmt.Errorf("笔记本标识符不能为空")
		logger.Error("笔记本标识符为空", zap.Error(err))
		fmt.Println("❌ 错误: 笔记本标识符不能为空")
		fmt.Println("💡 使用方法: siyuan-cli notebook getconf <笔记本名称或ID>")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
		return err
	}

	if opts.OutputFile != "" && !output.IsJSON("") {
		fmt.Println("⚠️  警告: 导出到文件时强制使用JSON格式")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		logger.Error("创建思源笔记客户端失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 创建思源笔记客户端失败: %v\n", err)
		return fmt.Errorf("创建客户端失败: %w", err)
	}
	logger.Info("思源笔记客户端创建成功")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("获取笔记本列表进行匹配")
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		logger.Error("获取笔记本列表失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		return fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	targetID, targetName, err := FindNotebook(notebooks, opts.NotebookIdentifier)
	if err != nil {
		logger.Error("笔记本匹配失败", zap.String("error", err.Error()), zap.String("identifier", opts.NotebookIdentifier))
		fmt.Printf("❌ %v\n", err)
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看所有可用的笔记本")
		return err
	}

	logger.Info("基于笔记本列表信息构建配置",
		zap.String("notebook_id", targetID),
		zap.String("notebook_name", targetName))

	var targetNotebook *siyuan.Notebook
	for _, nb := range notebooks {
		if nb.ID == targetID {
			targetNotebook = &nb
			break
		}
	}

	if targetNotebook == nil {
		err := fmt.Errorf("笔记本 '%s' 在列表中未找到", targetName)
		logger.Error("笔记本在列表中未找到", zap.String("notebook_name", targetName), zap.Error(err))
		return err
	}

	conf := &siyuan.NotebookConf{
		ID:       targetNotebook.ID,
		Name:     targetNotebook.Name,
		Sort:     targetNotebook.Sort,
		Icon:     targetNotebook.Icon,
		IconType: 0,
		Closed:   targetNotebook.Closed,
		SortMode: 0,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	logger.Info("成功构建笔记本配置",
		zap.String("notebook_id", targetID),
		zap.String("notebook_name", targetName))

	// JSON 格式或导出文件
	if output.IsJSON("") || opts.OutputFile != "" {
		return outputConfAsJSON(conf, opts.OutputFile)
	}

	// 默认：TUI 显示
	return outputConfAsTUI(conf)
}

// outputConfAsTUI 以TUI形式显示笔记本配置
func outputConfAsTUI(conf *siyuan.NotebookConf) error {
	t := ui.NewNotebookConfTUI(conf)
	return t.Run()
}

// outputConfAsJSON 以JSON格式输出笔记本配置
func outputConfAsJSON(conf *siyuan.NotebookConf, outputFile string) error {
	if outputFile != "" {
		jsonData, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON序列化失败: %w", err)
		}
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
		fmt.Printf("已导出到: %s\n", outputFile)
		return nil
	}

	output.PrintJSON(conf)
	return nil
}
