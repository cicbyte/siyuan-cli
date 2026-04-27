package ui

import (
	"fmt"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PathItem 表示路径选择器中的一个项目
type PathItem struct {
	HumanPath string // 人类可读路径
	Level     int    // 层级深度
	IsExpanded bool  // 是否展开（仅文件夹有）
	IsFolder  bool   // 是否为文件夹
	Parent    *PathItem
	Children  []*PathItem
}

// DocumentPathSelectorTUI 表示文档路径选择器的交互式界面
type DocumentPathSelectorTUI struct {
	notebook      string
	notebookID    string
	items         []*PathItem          // 显示项目
	rootItems     []*PathItem          // 根节点列表
	allItems      map[string]*PathItem // 所有项目的映射
	cursor        int                  // 当前选中的项目索引
	scroll        int                  // 滚动偏移量
	nameMap       map[string]string    // docID → hpath 映射
	docTree       *siyuan.DocTreeData
	quitting      bool
	selectedPath  string // 用户选择的路径
}

// NewDocumentPathSelectorTUI 创建新的文档路径选择器TUI
func NewDocumentPathSelectorTUI(notebook, notebookID string, docTree *siyuan.DocTreeData, nameMap map[string]string) *DocumentPathSelectorTUI {
	tui := &DocumentPathSelectorTUI{
		notebook:     notebook,
		notebookID:   notebookID,
		nameMap:      nameMap,
		docTree:      docTree,
		items:        make([]*PathItem, 0),
		rootItems:    make([]*PathItem, 0),
		allItems:     make(map[string]*PathItem),
		cursor:       0,
		scroll:       0,
		quitting:     false,
		selectedPath: "",
	}

	// 构建树形结构
	tui.buildPathTree()

	return tui
}

// buildPathTree 构建路径选择树
func (m *DocumentPathSelectorTUI) buildPathTree() {
	var allPaths []string
	pathSet := make(map[string]bool)

	allPaths = append(allPaths, "/")
	pathSet["/"] = true

	for _, hpath := range m.nameMap {
		if hpath != "" && !pathSet[hpath] {
			allPaths = append(allPaths, hpath)
			pathSet[hpath] = true

			parentPath := getParentPath(hpath)
			for parentPath != "" && parentPath != "/" && !pathSet[parentPath] {
				allPaths = append(allPaths, parentPath)
				pathSet[parentPath] = true
				parentPath = getParentPath(parentPath)
			}
		}
	}

	// 构建路径项目
	m.items = make([]*PathItem, 0, len(allPaths))
	m.rootItems = make([]*PathItem, 0)
	m.allItems = make(map[string]*PathItem)

	// 首先创建所有项目
	for _, path := range allPaths {
		item := &PathItem{
			HumanPath:  path,
			Level:      calculateLevelFromPath(path),
			IsExpanded: false,
			IsFolder:   true, // 在路径选择器中，所有项目都可以作为文件夹来创建文档
			Parent:     nil,
			Children:   make([]*PathItem, 0),
		}
		m.items = append(m.items, item)
		m.allItems[path] = item
	}

	// 构建父子关系
	for _, item := range m.items {
		path := item.HumanPath
		parentPath := getParentPath(path)
		if parentPath != "" {
			if parent, exists := m.allItems[parentPath]; exists {
				item.Parent = parent
				parent.Children = append(parent.Children, item)
				item.Level = parent.Level + 1
			}
		}
	}

	// 识别根节点
	for _, item := range m.items {
		if item.Parent == nil {
			m.rootItems = append(m.rootItems, item)
			// 默认展开第一层
			item.IsExpanded = true
			for _, child := range item.Children {
				m.items = append(m.items, child)
			}
		}
	}

	// 重新构建显示列表
	m.buildDisplayListFromExpandedItems()
}

// buildDisplayListFromExpandedItems 根据展开状态构建显示列表
func (m *DocumentPathSelectorTUI) buildDisplayListFromExpandedItems() {
	m.items = make([]*PathItem, 0)

	// 递归添加节点
	for _, root := range m.rootItems {
		m.addPathItemToList(root)
	}
}

// addPathItemToList 递归将节点添加到显示列表
func (m *DocumentPathSelectorTUI) addPathItemToList(item *PathItem) {
	m.items = append(m.items, item)

	if item.IsExpanded && len(item.Children) > 0 {
		for _, child := range item.Children {
			m.addPathItemToList(child)
		}
	}
}

// Init 初始化TUI
func (m *DocumentPathSelectorTUI) Init() tea.Cmd {
	return nil
}

// Update 更新TUI状态
func (m *DocumentPathSelectorTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "home", "g":
			m.cursor = 0
			m.scroll = 0
		case "end", "G":
			m.cursor = len(m.items) - 1
			m.adjustScroll()
		case "enter", " ":
			m.selectPath()
			m.quitting = true
			return m, tea.Quit
		case "r":
			// 刷新功能
			m.buildPathTree()
		}
	}
	return m, nil
}

// selectPath 选择当前路径
func (m *DocumentPathSelectorTUI) selectPath() {
	if m.cursor < len(m.items) {
		m.selectedPath = m.items[m.cursor].HumanPath
	}
}

// GetSelectedPath 获取用户选择的路径
func (m *DocumentPathSelectorTUI) GetSelectedPath() string {
	return m.selectedPath
}

// adjustScroll 调整滚动位置
func (m *DocumentPathSelectorTUI) adjustScroll() {
	height := 20

	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+height {
		m.scroll = m.cursor - height + 1
	}

	if m.scroll < 0 {
		m.scroll = 0
	}
	if m.scroll > len(m.items)-1 {
		m.scroll = len(m.items) - 1
	}
}

// View 渲染TUI界面
func (m *DocumentPathSelectorTUI) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207")).
		Align(lipgloss.Center).
		Width(80)

	title := titleStyle.Render("📍 选择文档保存位置 - " + m.notebook)

	// 渲染路径项目
	var pathContent strings.Builder
	height := 20

	start := m.scroll
	end := start + height
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := start; i < end; i++ {
		if i >= len(m.items) {
			break
		}

		item := m.items[i]
		line := m.renderPathItem(item, i == m.cursor)
		pathContent.WriteString(line)
		pathContent.WriteString("\n")
	}

	// 统计信息
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(80)

	statsText := fmt.Sprintf("📊 可选路径 %d 个 | %s", len(m.items), m.getSelectedPathDisplay())
	stats := statsStyle.Render(statsText)

	// 帮助信息
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(80)

	helpText := "↑↓ 选择 | Home/End 跳转 | Enter/空格 选择 | R 刷新 | Q 取消"
	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		stats,
		"",
		pathContent.String(),
		"",
		help,
	)
}

// getSelectedPathDisplay 获取选中路径的显示文本
func (m *DocumentPathSelectorTUI) getSelectedPathDisplay() string {
	if m.selectedPath != "" {
		return "已选择: " + m.selectedPath
	}
	if m.cursor < len(m.items) {
		return "当前: " + m.items[m.cursor].HumanPath
	}
	return ""
}

// renderPathItem 渲染单个路径项目
func (m *DocumentPathSelectorTUI) renderPathItem(item *PathItem, isSelected bool) string {
	// 构建缩进
	indent := strings.Repeat("  ", item.Level)

	// 构建图标
	var icon string
	if item.HumanPath == "/" {
		icon = "🏠"
	} else {
		icon = "📁"
	}

	// 构建显示文本
	displayText := item.HumanPath
	if displayText == "" {
		displayText = "/"
	}

	// 截断过长的文本
	maxLen := 50
	if len(displayText) > maxLen {
		displayText = "..." + displayText[len(displayText)-maxLen+3:]
	}

	line := fmt.Sprintf("%s%s %s", indent, icon, displayText)

	// 应用选中样式
	if isSelected {
		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("207")).
			Bold(true).
			Background(lipgloss.Color("236"))
		line = selectedStyle.Render(line)
	}

	return line
}

// Run 运行TUI
func (m *DocumentPathSelectorTUI) Run() (string, error) {
	program := tea.NewProgram(m)
	model, err := program.Run()
	if err != nil {
		return "", err
	}

	if tui, ok := model.(*DocumentPathSelectorTUI); ok {
		return tui.GetSelectedPath(), nil
	}

	return "", fmt.Errorf("TUI运行失败")
}

// getParentPath 获取父级路径
func getParentPath(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	cleanPath := strings.Trim(path, "/")
	lastSlash := strings.LastIndex(cleanPath, "/")
	if lastSlash == -1 {
		return ""
	}
	return "/" + cleanPath[:lastSlash]
}

// calculateLevelFromPath 从路径计算层级
func calculateLevelFromPath(path string) int {
	if path == "" {
		return 0
	}
	return strings.Count(strings.Trim(path, "/"), "/")
}