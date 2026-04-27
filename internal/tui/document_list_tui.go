package ui

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DocumentListTUI 表示文档列表的交互式界面
type DocumentListTUI struct {
	table    table.Model
	notebook string
	docTree  *siyuan.DocTreeData
	nameMap  map[string]string
	quitting bool
}

// NewDocumentListTUI 创建新的文档列表TUI
func NewDocumentListTUI(notebook string, notebookID string, docTree *siyuan.DocTreeData, nameMap map[string]string) *DocumentListTUI {
	columns := []table.Column{
		{Title: "文档路径", Width: 50},
		{Title: "文档ID", Width: 20},
	}

	rows := make([]table.Row, 0, len(nameMap))
	for docID, hpath := range nameMap {
		displayPath := hpath
		if displayPath == "" {
			displayPath = docID
		}
		if len(displayPath) > 45 {
			displayPath = "..." + displayPath[len(displayPath)-45:]
		}
		idDisplay := docID
		if len(idDisplay) > 18 {
			idDisplay = idDisplay[:15] + "..."
		}
		rows = append(rows, table.Row{displayPath, idDisplay})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("207")).
		Bold(true)

	s.Cell = s.Cell.
		BorderStyle(lipgloss.HiddenBorder())

	t.SetStyles(s)

	return &DocumentListTUI{
		table:    t,
		notebook: notebook,
		docTree:  docTree,
		nameMap:  nameMap,
	}
}

// Init 初始化TUI
func (m *DocumentListTUI) Init() tea.Cmd {
	return nil
}

// Update 更新TUI状态
func (m *DocumentListTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		case "home", "g":
			m.table.GotoTop()
		case "end", "G":
			m.table.GotoBottom()
		case "r":
			return m, nil
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View 渲染TUI界面
func (m *DocumentListTUI) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207")).
		Align(lipgloss.Center).
		Width(100)

	title := titleStyle.Render("📁 文档树 - " + m.notebook)

	var statsText string
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(100)

	if len(m.docTree.Tree) > 0 {
		totalNodes := countTreeNodes(m.docTree.Tree)
		statsText = fmt.Sprintf("📊 统计: 文档节点 %d 个 | Tree 结构 | q 退出", totalNodes)
	} else {
		statsText = "📊 统计: 暂无数据 | q 退出"
	}
	statsRendered := statsStyle.Render(statsText)

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Bold(true).
		Padding(0, 1)

	pathText := ""
	if m.docTree.Path != "" {
		pathText = pathStyle.Render("📍 路径: " + m.docTree.Path)
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(100)

	helpText := "↑↓ 选择 | Home/End 跳转 | Q 退出"
	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		pathText,
		"",
		m.table.View(),
		"",
		statsRendered,
		help,
	)
}

// Run 运行TUI
func (m *DocumentListTUI) Run() error {
	program := tea.NewProgram(m)
	_, err := program.Run()
	return err
}

func countTreeNodes(nodes []siyuan.DocTreeNode) int {
	count := len(nodes)
	for _, node := range nodes {
		if len(node.Children) > 0 {
			count += countTreeNodes(node.Children)
		}
	}
	return count
}
