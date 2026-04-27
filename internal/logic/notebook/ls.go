package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"go.uber.org/zap"
)

// Options 定义notebook list命令的选项
type Options struct {
	ShowClosed bool   // 是否显示已关闭的笔记本
	SortBy     string // 排序方式：name, id, created, updated
	OutputFile string // 输出文件路径
}

// ListNotebooks 执行列出笔记本的逻辑
func ListNotebooks(opts Options) error {
	logger := log.GetLogger()
	logger.Info("开始获取笔记本列表")

	if !siyuan.IsSiYuanConfigValid() {
		logger.Error("思源笔记配置无效或未启用")
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli auth login' 配置连接")
		return fmt.Errorf("思源笔记配置无效")
	}
	logger.Info("思源笔记配置验证通过")

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		logger.Error("创建思源笔记客户端失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 创建思源笔记客户端失败: %v\n", err)
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		logger.Error("获取笔记本列表失败", zap.String("error", err.Error()))
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		fmt.Println("\n连接诊断:")
		fmt.Printf("  - 请确认思源笔记是否正在运行\n")
		fmt.Printf("  - 请确认思源笔记的监听地址是否正确 (当前: %s)\n", common.GetAppConfig().SiYuan.BaseURL)
		if common.GetAppConfig().SiYuan.ApiToken == "" {
			fmt.Printf("  - API Token为空，如果思源笔记启用了访问授权，请在配置中设置 api_token\n")
		} else {
			fmt.Printf("  - API Token已配置\n")
		}
		fmt.Printf("  - 请确认网络连接是否正常\n")
		return fmt.Errorf("获取笔记本列表失败: %w", err)
	}
	logger.Info("成功获取笔记本列表", zap.Int("count", len(notebooks)))

	filteredNotebooks := filterNotebooks(notebooks, opts.ShowClosed)
	if len(filteredNotebooks) == 0 {
		fmt.Println("暂无笔记本")
		if !opts.ShowClosed {
			fmt.Println("提示: 使用 --closed 参数可以查看已关闭的笔记本")
		}
		return nil
	}

	sortedNotebooks := sortNotebooks(filteredNotebooks, opts.SortBy)

	if output.IsJSON("") {
		return outputJSON(sortedNotebooks, opts.OutputFile)
	}

	status := func(nb siyuan.Notebook) string {
		if nb.Closed {
			return "已关闭"
		}
		return "已打开"
	}

	headers := []string{"名称", "ID", "图标", "状态", "排序"}
	var rows [][]string
	for _, nb := range sortedNotebooks {
		rows = append(rows, []string{
			output.Truncate(nb.Name, 30),
			nb.ID,
			nb.Icon,
			status(nb),
			fmt.Sprintf("%d", nb.Sort),
		})
	}
	output.PrintTableRight(headers, rows, 5)
	fmt.Printf("\n共 %d 个笔记本\n", len(sortedNotebooks))
	return nil
}

func filterNotebooks(notebooks []siyuan.Notebook, showClosed bool) []siyuan.Notebook {
	var filtered []siyuan.Notebook
	for _, nb := range notebooks {
		if !showClosed && nb.Closed {
			continue
		}
		filtered = append(filtered, nb)
	}
	return filtered
}

func sortNotebooks(notebooks []siyuan.Notebook, sortBy string) []siyuan.Notebook {
	sorted := make([]siyuan.Notebook, len(notebooks))
	copy(sorted, notebooks)

	switch sortBy {
	case "id":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].ID < sorted[j].ID
		})
	case "sort":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Sort < sorted[j].Sort
		})
	case "name":
		fallthrough
	default:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
	}
	return sorted
}

func outputJSON(notebooks []siyuan.Notebook, filename string) error {
	type nbItem struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Icon   string `json:"icon"`
		Sort   int    `json:"sort"`
		Closed bool   `json:"closed"`
	}
	items := make([]nbItem, len(notebooks))
	for i, nb := range notebooks {
		items[i] = nbItem{ID: nb.ID, Name: nb.Name, Icon: nb.Icon, Sort: nb.Sort, Closed: nb.Closed}
	}

	data := map[string]any{"count": len(items), "list": items}
	if filename != "" {
		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON 编码失败: %w", err)
		}
		if err := os.WriteFile(filename, out, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
		fmt.Printf("已导出到: %s\n", filename)
		return nil
	}

	output.PrintJSON(data)
	return nil
}
