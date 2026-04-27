package ui

import (
	"fmt"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotebookConfTUI 表示笔记本配置的交互式界面
type NotebookConfTUI struct {
	table table.Model
	conf  *siyuan.NotebookConf
	quitting bool
}

// NewNotebookConfTUI 创建新的笔记本配置TUI
func NewNotebookConfTUI(conf *siyuan.NotebookConf) *NotebookConfTUI {
	// 定义表格样式
	columns := []table.Column{
		{Title: "配置项", Width: 20},
		{Title: "值", Width: 50},
	}

	// 准备配置数据
	rows := make([]table.Row, 0)

	// 基本信息
	rows = append(rows, table.Row{"", "📋 基本信息"})
	rows = append(rows, table.Row{"名称", conf.Name})
	rows = append(rows, table.Row{"ID", conf.ID})
	rows = append(rows, table.Row{"图标", getIconDisplay(conf.Icon, conf.IconType)})
	rows = append(rows, table.Row{"状态", getClosedStatus(conf.Closed)})

	rows = append(rows, table.Row{"", ""}) // 分隔行

	// 排序设置
	rows = append(rows, table.Row{"", "🔧 排序设置"})
	rows = append(rows, table.Row{"排序顺序", fmt.Sprintf("%d", conf.Sort)})
	rows = append(rows, table.Row{"排序模式", getSortModeDisplay(conf.SortMode)})

	rows = append(rows, table.Row{"", ""}) // 分隔行

	// 时间信息
	rows = append(rows, table.Row{"", "⏰ 时间信息"})
	rows = append(rows, table.Row{"创建时间", conf.Created.Format("2006-01-02 15:04:05")})
	rows = append(rows, table.Row{"更新时间", conf.Updated.Format("2006-01-02 15:04:05")})

	// 高级配置（如果有）
	hasAdvanced := false
	advancedRows := []table.Row{}

	if conf.Avatar != "" {
		if !hasAdvanced {
			advancedRows = append(advancedRows, table.Row{"", "🎛️ 高级配置"})
			hasAdvanced = true
		}
		advancedRows = append(advancedRows, table.Row{"头像", conf.Avatar})
	}

	if conf.Algo != "" {
		if !hasAdvanced {
			advancedRows = append(advancedRows, table.Row{"", "🎛️ 高级配置"})
			hasAdvanced = true
		}
		advancedRows = append(advancedRows, table.Row{"排序算法", conf.Algo})
	}

	if conf.RefCreateAnchor != "" {
		if !hasAdvanced {
			advancedRows = append(advancedRows, table.Row{"", "🎛️ 高级配置"})
			hasAdvanced = true
		}
		advancedRows = append(advancedRows, table.Row{"引用创建锚点", conf.RefCreateAnchor})
	}

	if conf.DocCreateSaveFolder != "" {
		if !hasAdvanced {
			advancedRows = append(advancedRows, table.Row{"", "🎛️ 高级配置"})
			hasAdvanced = true
		}
		advancedRows = append(advancedRows, table.Row{"文档保存文件夹", conf.DocCreateSaveFolder})
	}

	// 如果有高级配置，添加到表格中
	if hasAdvanced {
		rows = append(rows, table.Row{"", ""}) // 分隔行
		rows = append(rows, advancedRows...)
	}

	// 创建表格
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false), // 配置显示不需要焦点
		table.WithHeight(15),
	)

	// 设置表格样式
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.NoColor{}).
		Bold(false)

	s.Cell = s.Cell.
		BorderStyle(lipgloss.HiddenBorder())

	t.SetStyles(s)

	return &NotebookConfTUI{
		table: t,
		conf:  conf,
	}
}

// Init 初始化TUI
func (m *NotebookConfTUI) Init() tea.Cmd {
	return nil
}

// Update 更新TUI状态
func (m *NotebookConfTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View 渲染TUI界面
func (m *NotebookConfTUI) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207")).
		Align(lipgloss.Center).
		Width(80)

	title := titleStyle.Render("📋 笔记本配置信息")

	// 处理行样式
	rows := m.table.Rows()
	styledRows := make([]string, 0, len(rows))

	for _, row := range rows {
		if strings.HasPrefix(row[0], "📋") || strings.HasPrefix(row[0], "🔧") ||
		   strings.HasPrefix(row[0], "⏰") || strings.HasPrefix(row[0], "🎛️") {
			// 分类标题样式
			categoryStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("14")).
				Width(70)
			if row[0] == "" {
				// 仅右侧有标题的情况
				styledRows = append(styledRows, categoryStyle.Render(row[1]))
			} else {
				// 双列标题
				styledRows = append(styledRows, categoryStyle.Render(row[0]+" "+row[1]))
			}
		} else if row[0] == "" {
			// 分隔行
			styledRows = append(styledRows, "")
		} else {
			// 普通行：左侧标签加粗，右侧值正常显示
			labelStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("153")).
				Width(20)
			valueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Width(50)
			styledRows = append(styledRows,
				lipgloss.JoinHorizontal(lipgloss.Left,
					labelStyle.Render(row[0]),
					valueStyle.Render(row[1]),
				))
		}
	}

	content := strings.Join(styledRows, "\n")

	// 创建边框样式
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		Width(84)

	contentBox := borderStyle.Render(content)

	// 底部提示
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(80)

	footer := footerStyle.Render("按 q 键退出")

	// 组装完整界面
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		contentBox,
		"",
		footer,
	)
}

// Run 运行TUI
func (m *NotebookConfTUI) Run() error {
	program := tea.NewProgram(m)
	_, err := program.Run()
	return err
}

// 辅助函数：获取图标显示
func getIconDisplay(icon string, iconType int) string {
	if icon == "" {
		return "-"
	}
	if iconType == 0 {
		// emoji图标
		return icon
	}
	return fmt.Sprintf("%s (type:%d)", icon, iconType)
}

// 辅助函数：获取关闭状态显示
func getClosedStatus(closed bool) string {
	if closed {
		return "📁 已关闭"
	}
	return "📖 打开中"
}

// 辅助函数：获取排序模式显示
func getSortModeDisplay(sortMode int) string {
	switch sortMode {
	case 0:
		return "按自定义排序"
	case 1:
		return "按文件名排序"
	case 2:
		return "按创建时间排序"
	case 3:
		return "按修改时间排序"
	default:
		return fmt.Sprintf("未知模式 (%d)", sortMode)
	}
}