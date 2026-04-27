package ui

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotebookTUI 表示笔记本的交互式界面
type NotebookTUI struct {
	table table.Model
	notebooks []siyuan.Notebook
}

// NewNotebookTUI 创建新的笔记本TUI
func NewNotebookTUI(notebooks []siyuan.Notebook) *NotebookTUI {
	// 定义表格样式
	columns := []table.Column{
		{Title: "状态", Width: 8},
		{Title: "名称", Width: 20},
		{Title: "ID", Width: 12},
		{Title: "图标", Width: 8},
		{Title: "排序", Width: 6},
	}

	// 准备表格行数据
	rows := make([]table.Row, 0, len(notebooks))
	for _, nb := range notebooks {
		status := "📂 打开"
		if nb.Closed {
			status = "📁 关闭"
		}

		id := nb.ID
		if len(id) > 10 {
			id = id[:10] + "..."
		}

		icon := nb.Icon
		if icon == "" {
			icon = "-"
		}

		rows = append(rows, table.Row{
			status,
			nb.Name,
			id,
			icon,
			fmt.Sprintf("%d", nb.Sort),
		})
	}

	// 创建表格
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+3), // +3 for header and footer
	)

	// 定义样式
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)
	t.SetStyles(s)

	return &NotebookTUI{
		table: t,
		notebooks: notebooks,
	}
}

// View 渲染界面
func (n *NotebookTUI) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207")).
		Width(60).
		Align(lipgloss.Center).
		Render("笔记本列表")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("共 %d 个笔记本 | ↑↓ 移动 | q 退出", len(n.notebooks)))

	// 统计信息
	openCount := 0
	closedCount := 0
	for _, nb := range n.notebooks {
		if nb.Closed {
			closedCount++
		} else {
			openCount++
		}
	}

	stats := lipgloss.NewStyle().
		Foreground(lipgloss.Color("144")).
		Render(fmt.Sprintf("📊 统计: 打开 %d 个，关闭 %d 个", openCount, closedCount))

	// 组合所有元素
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		n.table.View(),
		"",
		stats,
		"",
		footer,
	)

	// 添加边框
	border := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2)

	return border.Render(content)
}

// Init 初始化模型
func (n *NotebookTUI) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (n *NotebookTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return n, tea.Quit
		}
	}

	n.table, cmd = n.table.Update(msg)
	return n, cmd
}

// GetSelectedNotebook 获取当前选中的笔记本
func (n *NotebookTUI) GetSelectedNotebook() *siyuan.Notebook {
	if len(n.notebooks) == 0 {
		return nil
	}

	selectedIndex := n.table.Cursor()
	if selectedIndex >= 0 && selectedIndex < len(n.notebooks) {
		return &n.notebooks[selectedIndex]
	}

	return nil
}