package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/cicbyte/siyuan-cli/internal/tui"
	"go.uber.org/zap"
)

// SetConfOptions 定义notebook setconf命令的选项
type SetConfOptions struct {
	NotebookIdentifier string // 笔记本ID或名称
	InputMode          string // 输入模式：cli, json, tui
	ConfigFile         string // JSON配置文件路径
	ConfigData         map[string]interface{} // CLI直接设置的配置数据
}

// SetNotebookConf 执行设置笔记本配置的逻辑
func SetNotebookConf(opts SetConfOptions) error {
	logger := log.GetLogger()
	logger.Info("开始设置笔记本配置",
		zap.String("identifier", opts.NotebookIdentifier),
		zap.String("input_mode", opts.InputMode),
		zap.String("config_file", opts.ConfigFile))

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
		fmt.Println("💡 使用方法: siyuan-cli notebook setconf <笔记本名称或ID>")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
		return err
	}

	// 验证输入模式
	validModes := []string{"cli", "json", "tui"}
	isValidMode := false
	for _, mode := range validModes {
		if opts.InputMode == mode {
			isValidMode = true
			break
		}
	}
	if !isValidMode {
		err := fmt.Errorf("无效的输入模式: %s，支持的模式: %s", opts.InputMode, strings.Join(validModes, ", "))
		logger.Error("无效的输入模式", zap.String("mode", opts.InputMode), zap.Error(err))
		fmt.Printf("❌ 错误: %v\n", err)
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

	// 获取当前配置作为基础
	logger.Info("获取当前笔记本配置")
	baseConf, err := getBaseConfig(ctx, client, targetID)
	if err != nil {
		logger.Error("获取基础配置失败", zap.String("error", err.Error()), zap.String("notebook_id", targetID))
		fmt.Printf("❌ 获取当前配置失败: %v\n", err)
		return fmt.Errorf("获取当前配置失败: %w", err)
	}

	// 根据输入模式处理配置
	var finalConf *siyuan.NotebookConf
	switch opts.InputMode {
	case "cli":
		finalConf, err = processCLIInput(baseConf, opts.ConfigData)
	case "json":
		finalConf, err = processJSONInput(baseConf, opts.ConfigFile)
	case "tui":
		finalConf, err = processTUIInput(baseConf)
	default:
		err = fmt.Errorf("不支持的输入模式: %s", opts.InputMode)
	}

	if err != nil {
		logger.Error("处理配置输入失败", zap.String("error", err.Error()), zap.String("input_mode", opts.InputMode))
		fmt.Printf("❌ 处理配置失败: %v\n", err)
		return err
	}

	// 应用配置
	logger.Info("开始应用笔记本配置",
		zap.String("notebook_id", targetID),
		zap.String("notebook_name", finalConf.Name))

	err = client.SetNotebookConf(ctx, targetID, finalConf)
	if err != nil {
		logger.Error("设置笔记本配置失败",
			zap.String("error", err.Error()),
			zap.String("notebook_id", targetID),
			zap.String("notebook_name", finalConf.Name))

		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		} else {
			fmt.Printf("❌ 设置笔记本配置失败: %v\n", err)
			fmt.Println("\n🔍 错误诊断:")
			fmt.Printf("   - 请确认思源笔记是否正在运行\n")
			fmt.Printf("   - 请确认笔记本 '%s' (%s) 是否存在\n", targetName, targetID)
			fmt.Printf("   - 请确认配置数据格式是否正确\n")
		}
		return fmt.Errorf("设置笔记本配置失败: %w", err)
	}

	logger.Info("成功设置笔记本配置",
		zap.String("notebook_id", targetID),
		zap.String("notebook_name", finalConf.Name))

	fmt.Println("成功设置笔记本配置")
	var rows [][]string
	rows = append(rows, []string{"笔记本", fmt.Sprintf("%s (%s)", finalConf.Name, targetID)})
	if finalConf.Name != baseConf.Name {
		rows = append(rows, []string{"名称", fmt.Sprintf("%s → %s", baseConf.Name, finalConf.Name)})
	}
	if finalConf.Icon != baseConf.Icon {
		rows = append(rows, []string{"图标", fmt.Sprintf("%s → %s", baseConf.Icon, finalConf.Icon)})
	}
	if finalConf.Closed != baseConf.Closed {
		rows = append(rows, []string{"状态", fmt.Sprintf("%v → %v", !baseConf.Closed, finalConf.Closed)})
	}
	output.PrintTable([]string{"属性", "值"}, rows)

	return nil
}

// getBaseConfig 获取基础配置
func getBaseConfig(ctx context.Context, client *siyuan.Client, notebookID string) (*siyuan.NotebookConf, error) {
	// 尝试从API获取配置
	conf, err := client.GetNotebookConf(ctx, notebookID)
	if err == nil {
		return conf, nil
	}

	// 如果API不可用，从笔记本列表获取基础信息
	logger := log.GetLogger()
	logger.Warn("API不可用，从笔记本列表获取基础配置", zap.String("error", err.Error()))

	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	for _, nb := range notebooks {
		if nb.ID == notebookID {
			return &siyuan.NotebookConf{
				ID:       nb.ID,
				Name:     nb.Name,
				Icon:     nb.Icon,
				Sort:     nb.Sort,
				Closed:   nb.Closed,
				SortMode: 0, // 默认值
				Created:  time.Now(),
				Updated:  time.Now(),
			}, nil
		}
	}

	return nil, fmt.Errorf("未找到笔记本: %s", notebookID)
}

// processCLIInput 处理CLI直接输入
func processCLIInput(baseConf *siyuan.NotebookConf, configData map[string]interface{}) (*siyuan.NotebookConf, error) {
	logger := log.GetLogger()
	logger.Info("处理CLI输入配置", zap.Any("config_data", configData))

	conf := *baseConf // 复制基础配置

	// 处理各个配置项
	if name, ok := configData["name"].(string); ok && name != "" {
		conf.Name = name
	}
	if icon, ok := configData["icon"].(string); ok {
		conf.Icon = icon
	}
	if sort, ok := configData["sort"].(int); ok {
		conf.Sort = sort
	} else if sortStr, ok := configData["sort"].(string); ok {
		if sort, err := strconv.Atoi(sortStr); err == nil {
			conf.Sort = sort
		}
	}
	if sortMode, ok := configData["sortMode"].(int); ok {
		conf.SortMode = sortMode
	} else if sortModeStr, ok := configData["sortMode"].(string); ok {
		if sortMode, err := strconv.Atoi(sortModeStr); err == nil {
			conf.SortMode = sortMode
		}
	}
	if closed, ok := configData["closed"].(bool); ok {
		conf.Closed = closed
	}
	if refCreateAnchor, ok := configData["refCreateAnchor"].(string); ok {
		conf.RefCreateAnchor = refCreateAnchor
	}
	if docCreateSaveFolder, ok := configData["docCreateSaveFolder"].(string); ok {
		conf.DocCreateSaveFolder = docCreateSaveFolder
	}

	return &conf, nil
}

// processJSONInput 处理JSON文件输入
func processJSONInput(baseConf *siyuan.NotebookConf, configFile string) (*siyuan.NotebookConf, error) {
	logger := log.GetLogger()
	logger.Info("处理JSON文件输入", zap.String("config_file", configFile))

	// 读取JSON文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析JSON
	var confData map[string]interface{}
	if err := json.Unmarshal(data, &confData); err != nil {
		return nil, fmt.Errorf("解析JSON配置文件失败: %w", err)
	}

	// 处理配置数据
	return processCLIInput(baseConf, confData)
}

// processTUIInput 处理TUI交互式输入
func processTUIInput(baseConf *siyuan.NotebookConf) (*siyuan.NotebookConf, error) {
	logger := log.GetLogger()
	logger.Info("启动TUI交互式配置")

	// 创建TUI
	tui := ui.NewNotebookSetConfTUI(baseConf)

	// 运行TUI
	conf, err := tui.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI交互失败: %w", err)
	}

	// 检查是否有修改
	if !tui.GetIsDirty() {
		return nil, fmt.Errorf("没有修改任何配置")
	}

	return conf, nil
}