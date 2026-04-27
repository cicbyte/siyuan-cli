package ui

import (
	"fmt"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TreeItem 表示树形结构中的一个项目
type TreeItem struct {
	ID          string      // 文档ID
	HumanPath   string      // 人类可读路径
	Level       int         // 层级深度
	IsExpanded  bool        // 是否展开
	HasChildren bool        // 是否有子节点
	Parent      *TreeItem   // 父节点
	Children    []*TreeItem // 子节点
}

// DocumentTreeTUI 表示文档树的交互式界面
type DocumentTreeTUI struct {
	notebook   string
	notebookID string
	items      []*TreeItem          // 扁平化的显示项目
	rootItems  []*TreeItem          // 根节点列表（保持原始顺序）
	allItems   map[string]*TreeItem // 所有项目的映射（按ID）
	cursor     int                  // 当前选中的项目索引
	scroll     int                  // 滚动偏移量
	nameMap    map[string]string    // docID → hpath 映射
	docTree    *siyuan.DocTreeData
	quitting   bool
}

// NewDocumentTreeTUI 创建新的文档树TUI
func NewDocumentTreeTUI(notebook, notebookID string, docTree *siyuan.DocTreeData, nameMap map[string]string) *DocumentTreeTUI {
	tui := &DocumentTreeTUI{
		notebook:   notebook,
		notebookID: notebookID,
		nameMap:    nameMap,
		docTree:    docTree,
		items:      make([]*TreeItem, 0),
		rootItems:  make([]*TreeItem, 0),
		allItems:   make(map[string]*TreeItem),
		cursor:     0,
		scroll:     0,
		quitting:   false,
	}

	// 构建树形结构
	tui.buildTree()

	return tui
}

// buildTree 构建树形结构
func (m *DocumentTreeTUI) buildTree() {
	if len(m.nameMap) > 0 && len(m.docTree.Tree) > 0 {
		m.buildTreeFromNameMap()
		return
	}

	m.items = make([]*TreeItem, 0)

	if len(m.docTree.Tree) > 0 {
		m.buildTreeFromNodes(m.docTree.Tree, 0, nil)
	} else if len(m.docTree.Files) > 0 {
		m.buildTreeFromFiles(m.docTree.Files)
	}
}

// buildTreeFromNameMap 从 nameMap 构建层级结构
func (m *DocumentTreeTUI) buildTreeFromNameMap() {
	m.items = make([]*TreeItem, 0, len(m.nameMap))
	m.rootItems = make([]*TreeItem, 0)
	m.allItems = make(map[string]*TreeItem)

	var allItems []*TreeItem
	pathToItem := make(map[string]*TreeItem)

	for docID, hpath := range m.nameMap {
		item := &TreeItem{
			ID:          docID,
			HumanPath:   hpath,
			Level:       calculateLevelFromPath(hpath),
			IsExpanded:  false,
			HasChildren: false,
			Parent:      nil,
			Children:    make([]*TreeItem, 0),
		}
		allItems = append(allItems, item)
		pathToItem[hpath] = item
		m.allItems[docID] = item
	}

	for _, item := range allItems {
		parentPath := getParentPath(item.HumanPath)
		if parentPath != "" {
			if parent, exists := pathToItem[parentPath]; exists {
				item.Parent = parent
				parent.Children = append(parent.Children, item)
				parent.HasChildren = true
				item.Level = parent.Level + 1
			}
		}
	}

	for _, item := range allItems {
		if item.Parent == nil {
			m.rootItems = append(m.rootItems, item)
			item.IsExpanded = true
			for _, child := range item.Children {
				m.items = append(m.items, child)
			}
		}
	}

	m.buildDisplayListFromExpandedItems()
}

// buildTreeFromNodes 从DocTreeNode构建树形结构
func (m *DocumentTreeTUI) buildTreeFromNodes(nodes []siyuan.DocTreeNode, level int, parent *TreeItem) {
	for _, node := range nodes {
		humanPath := m.nameMap[node.ID]
		if humanPath == "" {
			humanPath = node.ID
		}

		hasChildren := len(node.Children) > 0

		item := &TreeItem{
			ID:          node.ID,
			HumanPath:   humanPath,
			Level:       level,
			IsExpanded:  level < 1,
			HasChildren: hasChildren,
			Parent:      parent,
			Children:    make([]*TreeItem, 0),
		}

		if hasChildren {
			for _, childNode := range node.Children {
				childHumanPath := m.nameMap[childNode.ID]
				if childHumanPath == "" {
					childHumanPath = childNode.ID
				}
				childItem := &TreeItem{
					ID:          childNode.ID,
					HumanPath:   childHumanPath,
					Level:       level + 1,
					IsExpanded:  false,
					HasChildren: len(childNode.Children) > 0,
					Parent:      item,
					Children:    make([]*TreeItem, 0),
				}
				item.Children = append(item.Children, childItem)
			}
		}

		m.items = append(m.items, item)

		if item.HasChildren && item.IsExpanded {
			m.buildTreeFromNodes(node.Children, level+1, item)
		}
	}
}

// buildTreeFromFiles 从DocFile构建树形结构（兼容性）
func (m *DocumentTreeTUI) buildTreeFromFiles(files []siyuan.DocFile) {
	for _, file := range files {
		humanPath := m.nameMap[file.ID]
		if humanPath == "" {
			humanPath = file.Path
			if humanPath == "" {
				humanPath = file.ID
			}
		}

		level := strings.Count(strings.Trim(file.Path, "/"), "/")
		if file.Path == "" {
			level = 0
		}

		item := &TreeItem{
			ID:          file.ID,
			HumanPath:   humanPath,
			Level:       level,
			IsExpanded:  false,
			HasChildren: false,
			Parent:      nil,
			Children:    make([]*TreeItem, 0),
		}

		m.items = append(m.items, item)
	}
}

// Init 初始化TUI
func (m *DocumentTreeTUI) Init() tea.Cmd {
	return nil
}

// Update 更新TUI状态
func (m *DocumentTreeTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.toggleExpand()
		case "r":
			// 刷新功能
			m.refreshTree()
		}
	}
	return m, nil
}

// toggleExpand 切换展开/收缩状态
func (m *DocumentTreeTUI) toggleExpand() {
	if m.cursor >= len(m.items) {
		return
	}

	item := m.items[m.cursor]
	if !item.HasChildren {
		return // 没有子节点，无需切换
	}

	// 切换展开状态
	item.IsExpanded = !item.IsExpanded

	// 重新构建显示列表
	m.rebuildDisplayList()
}

// rebuildDisplayList 重新构建显示列表
func (m *DocumentTreeTUI) rebuildDisplayList() {
	// 保存当前的展开状态
	expandedStates := make(map[string]bool)
	for _, item := range m.items {
		if item.HasChildren {
			expandedStates[item.ID] = item.IsExpanded
		}
	}

	// 恢复展开状态到所有项目
	for _, item := range m.allItems {
		if isExpanded, exists := expandedStates[item.ID]; exists {
			item.IsExpanded = isExpanded
		}
	}

	// 根据展开状态重新构建显示列表（保持原始顺序）
	m.buildDisplayListFromExpandedItems()
}

// buildDisplayListFromExpandedItems 根据展开状态构建显示列表
func (m *DocumentTreeTUI) buildDisplayListFromExpandedItems() {
	// 使用保存的根节点列表（保持原始顺序）
	m.items = make([]*TreeItem, 0)

	// 递归添加节点，按照根节点的原始顺序
	for _, root := range m.rootItems {
		m.addItemToList(root)
	}
}

// addItemToList 递归将节点添加到显示列表
func (m *DocumentTreeTUI) addItemToList(item *TreeItem) {
	// 添加当前节点
	m.items = append(m.items, item)

	// 如果节点已展开且有子节点，递归添加子节点
	if item.IsExpanded && len(item.Children) > 0 {
		for _, child := range item.Children {
			m.addItemToList(child)
		}
	}
}

// buildDisplayListFromNodes 递归构建显示列表
func (m *DocumentTreeTUI) buildDisplayListFromNodes(nodes []siyuan.DocTreeNode, level int, parent *TreeItem) {
	for _, node := range nodes {
		humanPath := m.nameMap[node.ID]
		if humanPath == "" {
			humanPath = node.ID
		}

		hasChildren := len(node.Children) > 0

		item := &TreeItem{
			ID:          node.ID,
			HumanPath:   humanPath,
			Level:       level,
			HasChildren: hasChildren,
			Parent:      parent,
			Children:    make([]*TreeItem, 0),
		}

		for _, existingItem := range m.items {
			if existingItem.ID == node.ID {
				item.IsExpanded = existingItem.IsExpanded
				break
			}
		}

		m.items = append(m.items, item)

		if hasChildren && item.IsExpanded {
			m.buildDisplayListFromNodes(node.Children, level+1, item)
		}
	}
}

// buildDisplayListFromFiles 构建文件显示列表（兼容性）
func (m *DocumentTreeTUI) buildDisplayListFromFiles(files []siyuan.DocFile) {
	for _, file := range files {
		humanPath := m.nameMap[file.ID]
		if humanPath == "" {
			humanPath = file.Path
			if humanPath == "" {
				humanPath = file.ID
			}
		}

		level := strings.Count(strings.Trim(file.Path, "/"), "/")
		if file.Path == "" {
			level = 0
		}

		item := &TreeItem{
			ID:          file.ID,
			HumanPath:   humanPath,
			Level:       level,
			HasChildren: false,
			Parent:      nil,
			Children:    make([]*TreeItem, 0),
		}

		m.items = append(m.items, item)
	}
}

// findItemIndex 查找项目在显示列表中的索引
func (m *DocumentTreeTUI) findItemIndex(id string) int {
	for i, item := range m.items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

// refreshTree 刷新树形结构
func (m *DocumentTreeTUI) refreshTree() {
	m.buildTree()
}

// adjustScroll 调整滚动位置
func (m *DocumentTreeTUI) adjustScroll() {
	// 简单的滚动逻辑，确保光标项可见
	// 这里假设显示高度为20行
	height := 20

	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+height {
		m.scroll = m.cursor - height + 1
	}

	// 确保滚动位置不越界
	if m.scroll < 0 {
		m.scroll = 0
	}
	if m.scroll > len(m.items)-1 {
		m.scroll = len(m.items) - 1
	}
}

// View 渲染TUI界面
func (m *DocumentTreeTUI) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207")).
		Align(lipgloss.Center).
		Width(100)

	title := titleStyle.Render("📁 文档树 - " + m.notebook)

	// 渲染树形项目
	var treeContent strings.Builder
	height := 20 // 显示高度

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
		line := m.renderTreeItem(item, i == m.cursor)
		treeContent.WriteString(line)
		treeContent.WriteString("\n")
	}

	// 统计信息
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(100)

	statsText := fmt.Sprintf("📊 统计: 项目 %d 个 | 可展开 %d 个 | Tree 结构 | q 退出",
		len(m.items), m.countExpandableItems())
	stats := statsStyle.Render(statsText)

	// 帮助信息
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(100)

	helpText := "↑↓ 选择 | Home/End 跳转 | Enter/空格 展开/收缩 | Q 退出"
	help := helpStyle.Render(helpText)

	// 当前选中项的详细信息
	var details string
	if m.cursor < len(m.items) {
		item := m.items[m.cursor]
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Padding(0, 1)

		details = detailStyle.Render(fmt.Sprintf("📍 选中: %s | ID: %s", item.HumanPath, item.ID[:12] + "..."))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		details,
		"",
		treeContent.String(),
		"",
		stats,
		help,
	)
}

// renderTreeItem 渲染单个树形项目
func (m *DocumentTreeTUI) renderTreeItem(item *TreeItem, isSelected bool) string {
	// 构建缩进
	indent := strings.Repeat("  ", item.Level)

	// 构建展开/收缩图标
	var icon string
	if item.HasChildren {
		if item.IsExpanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
	} else {
		icon = "📄"
	}

	// 构建显示文本
	displayText := item.HumanPath
	if displayText == "" {
		displayText = item.ID
	}

	// 截断过长的文本
	maxLen := 60
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

// countExpandableItems 统计可展开的项目数量
func (m *DocumentTreeTUI) countExpandableItems() int {
	count := 0
	for _, item := range m.items {
		if item.HasChildren {
			count++
		}
	}
	return count
}

// Run 运行TUI
func (m *DocumentTreeTUI) Run() error {
	program := tea.NewProgram(m)
	_, err := program.Run()
	return err
}