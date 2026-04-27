package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotebookSetConfTUI 表示笔记本配置设置的交互式界面
type NotebookSetConfTUI struct {
	table   table.Model
	inputs  []textinput.Model
	current int
	editing bool
	conf    *siyuan.NotebookConf
	quitting bool
	dirty   bool
}

// SetConfField 表示配置字段
type SetConfField struct {
	Name        string
	Description string
	Type        string // "string", "int", "bool", "select"
	Value       interface{}
	Options     []string // 用于select类型
}

// NewNotebookSetConfTUI 创建新的笔记本配置设置TUI
func NewNotebookSetConfTUI(conf *siyuan.NotebookConf) *NotebookSetConfTUI {
	// 定义可配置字段
	fields := []SetConfField{
		{
			Name:        "name",
			Description: "笔记本名称",
			Type:        "string",
			Value:       conf.Name,
		},
		{
			Name:        "icon",
			Description: "笔记本图标",
			Type:        "string",
			Value:       conf.Icon,
		},
		{
			Name:        "sort",
			Description: "排序顺序",
			Type:        "int",
			Value:       conf.Sort,
		},
		{
			Name:        "sortMode",
			Description: "排序模式",
			Type:        "select",
			Value:       conf.SortMode,
			Options:     []string{"0: 自定义排序", "1: 按文件名排序", "2: 按创建时间排序", "3: 按修改时间排序"},
		},
		{
			Name:        "closed",
			Description: "是否关闭",
			Type:        "bool",
			Value:       conf.Closed,
		},
		{
			Name:        "refCreateAnchor",
			Description: "引用创建锚点",
			Type:        "string",
			Value:       conf.RefCreateAnchor,
		},
		{
			Name:        "docCreateSaveFolder",
			Description: "文档保存文件夹",
			Type:        "string",
			Value:       conf.DocCreateSaveFolder,
		},
	}

	// 创建表格
	columns := []table.Column{
		{Title: "配置项", Width: 25},
		{Title: "当前值", Width: 20},
		{Title: "说明", Width: 35},
	}

	rows := make([]table.Row, 0, len(fields))
	for _, field := range fields {
		valueStr := formatFieldValue(field.Value, field.Type)
		rows = append(rows, table.Row{
			field.Name,
			valueStr,
			field.Description,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
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

	// 创建文本输入框
	inputs := make([]textinput.Model, len(fields))
	for i := range fields {
		input := textinput.New()
		input.Placeholder = "输入新值..."
		input.CharLimit = 156
		input.Width = 50
		input.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("207")).Render("> ")
		input.Blur()
		inputs[i] = input
	}

	return &NotebookSetConfTUI{
		table:   t,
		inputs:  inputs,
		current: 0,
		conf:    conf,
	}
}

// Init 初始化TUI
func (m *NotebookSetConfTUI) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新TUI状态
func (m *NotebookSetConfTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			switch msg.String() {
			case "enter":
				m.saveCurrentField()
				m.editing = false
				m.inputs[m.current].Blur()
				return m, nil
			case "esc":
				m.editing = false
				m.inputs[m.current].Blur()
				return m, nil
			default:
				m.inputs[m.current], cmd = m.inputs[m.current].Update(msg)
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				if m.current > 0 {
					m.current--
				}
				m.table.SetCursor(m.current)
			case "down", "j":
				if m.current < len(m.inputs)-1 {
					m.current++
				}
				m.table.SetCursor(m.current)
			case "enter", " ":
				m.editing = true
				m.inputs[m.current].Focus()
				m.inputs[m.current].SetValue("")
				cmd = textinput.Blink
			case "s", "S":
				// 保存所有配置并退出
				m.saveAllFields()
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	// 更新表格高亮
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View 渲染TUI界面
func (m *NotebookSetConfTUI) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207")).
		Align(lipgloss.Center).
		Width(80)

	title := titleStyle.Render("⚙️ 笔记本配置设置")

	// 表格视图
	tableView := m.table.View()

	// 输入框视图
	var inputView string
	if m.editing {
		inputView = lipgloss.NewStyle().
			Padding(0, 2).
			Render(m.inputs[m.current].View())
	} else {
		inputView = ""
	}

	// 帮助信息
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(80)

	helpText := "↑↓ 选择 | Enter 编辑 | S 保存并退出 | Q 退出"
	if m.editing {
		helpText = "Enter 保存 | Esc 取消编辑 | Q 退出"
	}

	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		tableView,
		inputView,
		"",
		help,
	)
}

// saveCurrentField 保存当前字段
func (m *NotebookSetConfTUI) saveCurrentField() {
	fieldName := getFieldByIndex(m.current)
	value := strings.TrimSpace(m.inputs[m.current].Value())

	switch fieldName {
	case "name":
		if value != "" {
			m.conf.Name = value
			m.dirty = true
		}
	case "icon":
		m.conf.Icon = value
		m.dirty = true
	case "sort":
		if value != "" {
			if val, err := strconv.Atoi(value); err == nil {
				m.conf.Sort = val
				m.dirty = true
			}
		}
	case "sortMode":
		if strings.HasPrefix(value, "0:") {
			m.conf.SortMode = 0
			m.dirty = true
		} else if strings.HasPrefix(value, "1:") {
			m.conf.SortMode = 1
			m.dirty = true
		} else if strings.HasPrefix(value, "2:") {
			m.conf.SortMode = 2
			m.dirty = true
		} else if strings.HasPrefix(value, "3:") {
			m.conf.SortMode = 3
			m.dirty = true
		}
	case "closed":
		lowerValue := strings.ToLower(value)
		if lowerValue == "true" || lowerValue == "1" || lowerValue == "yes" || lowerValue == "y" {
			m.conf.Closed = true
			m.dirty = true
		} else if lowerValue == "false" || lowerValue == "0" || lowerValue == "no" || lowerValue == "n" {
			m.conf.Closed = false
			m.dirty = true
		}
	case "refCreateAnchor":
		m.conf.RefCreateAnchor = value
		m.dirty = true
	case "docCreateSaveFolder":
		m.conf.DocCreateSaveFolder = value
		m.dirty = true
	}

	// 更新表格显示
	m.updateTable()
}

// saveAllFields 保存所有字段
func (m *NotebookSetConfTUI) saveAllFields() {
	// 保存所有输入框的值
	for i := range m.inputs {
		m.current = i
		value := strings.TrimSpace(m.inputs[i].Value())
		if value != "" {
			m.saveCurrentField()
		}
	}
}

// updateTable 更新表格内容
func (m *NotebookSetConfTUI) updateTable() {
	fields := []SetConfField{
		{Name: "name", Description: "笔记本名称", Type: "string", Value: m.conf.Name},
		{Name: "icon", Description: "笔记本图标", Type: "string", Value: m.conf.Icon},
		{Name: "sort", Description: "排序顺序", Type: "int", Value: m.conf.Sort},
		{Name: "sortMode", Description: "排序模式", Type: "select", Value: m.conf.SortMode,
			Options: []string{"0: 自定义排序", "1: 按文件名排序", "2: 按创建时间排序", "3: 按修改时间排序"}},
		{Name: "closed", Description: "是否关闭", Type: "bool", Value: m.conf.Closed},
		{Name: "refCreateAnchor", Description: "引用创建锚点", Type: "string", Value: m.conf.RefCreateAnchor},
		{Name: "docCreateSaveFolder", Description: "文档保存文件夹", Type: "string", Value: m.conf.DocCreateSaveFolder},
	}

	rows := make([]table.Row, 0, len(fields))
	for i, field := range fields {
		valueStr := formatFieldValue(field.Value, field.Type)
		if i == m.current && m.editing {
			valueStr += " [编辑中...]"
		}
		rows = append(rows, table.Row{
			field.Name,
			valueStr,
			field.Description,
		})
	}

	m.table.SetRows(rows)
}

// Run 运行TUI
func (m *NotebookSetConfTUI) Run() (*siyuan.NotebookConf, error) {
	program := tea.NewProgram(m)
	model, err := program.Run()
	if err != nil {
		return nil, err
	}

	if tui, ok := model.(*NotebookSetConfTUI); ok {
		if tui.dirty {
			return tui.conf, nil
		}
	}
	return nil, fmt.Errorf("没有修改任何配置")
}

// GetIsDirty 返回是否修改了配置
func (m *NotebookSetConfTUI) GetIsDirty() bool {
	return m.dirty
}

// 辅助函数：格式化字段值
func formatFieldValue(value interface{}, fieldType string) string {
	switch fieldType {
	case "bool":
		if v, ok := value.(bool); ok {
			if v {
				return "是"
			}
			return "否"
		}
	case "select":
		if v, ok := value.(int); ok {
			options := []string{"自定义排序", "按文件名排序", "按创建时间排序", "按修改时间排序"}
			if v >= 0 && v < len(options) {
				return options[v]
			}
		}
	default:
		return fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("%v", value)
}

// 辅助函数：根据索引获取字段名
func getFieldByIndex(index int) string {
	fields := []string{"name", "icon", "sort", "sortMode", "closed", "refCreateAnchor", "docCreateSaveFolder"}
	if index >= 0 && index < len(fields) {
		return fields[index]
	}
	return ""
}