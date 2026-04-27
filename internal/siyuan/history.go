package siyuan

import (
	"context"
	"strings"
)

// OutlineItem 大纲项
type OutlineItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Level   int    `json:"-"`
	SubType string `json:"subType"`
	Depth   int    `json:"depth"`
}

type outlineRaw struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	SubType string        `json:"subType"`
	Depth   int           `json:"depth"`
	Blocks  []outlineBlock `json:"blocks,omitempty"`
}

type outlineBlock struct {
	ID       string        `json:"id"`
	Content  string        `json:"content"`
	SubType  string        `json:"subType"`
	Children []outlineBlock `json:"children,omitempty"`
}

// CreateDailyNoteOptions 创建日记选项
type CreateDailyNoteOptions struct {
	Notebook string `json:"notebook"`
	App      string `json:"app,omitempty"`
}

// CreateDailyNoteResponse 创建日记响应
type CreateDailyNoteResponse struct {
	ID     string `json:"id"`
	RootID string `json:"rootID"`
	Box    string `json:"box"`
	Path   string `json:"path"`
	HPath  string `json:"hPath"`
	Name   string `json:"name"`
}

// CreateDailyNote 创建日记
// 路由：POST /api/filetree/createDailyNote
func (c *Client) CreateDailyNote(ctx context.Context, opts CreateDailyNoteOptions) (*CreateDailyNoteResponse, error) {
	var ret CreateDailyNoteResponse
	if err := c.post(ctx, "/api/filetree/createDailyNote", opts, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// GetDocOutline 获取文档大纲
// 路由：POST /api/outline/getDocOutline
func (c *Client) GetDocOutline(ctx context.Context, id string) ([]OutlineItem, error) {
	var raw []outlineRaw
	if err := c.post(ctx, "/api/outline/getDocOutline", map[string]string{"id": id}, &raw); err != nil {
		return nil, err
	}
	return flattenOutline(raw), nil
}

func flattenOutline(items []outlineRaw) []OutlineItem {
	var result []OutlineItem
	for _, item := range items {
		content := strings.ReplaceAll(item.Name, "&nbsp;", " ")
		if content == "" {
			content = item.SubType
		}
		result = append(result, OutlineItem{
			ID:      item.ID,
			Content: content,
			Level:   levelFromSubType(item.SubType),
			Depth:   item.Depth,
		})
		for _, block := range item.Blocks {
			result = append(result, flattenBlock(block, item.Depth+1)...)
		}
	}
	return result
}

func flattenBlock(block outlineBlock, baseDepth int) []OutlineItem {
	level := levelFromSubType(block.SubType)
	if level == 0 {
		level = 1
	}
	items := []OutlineItem{{
		ID:      block.ID,
		Content: strings.ReplaceAll(block.Content, "&nbsp;", " "),
		Level:   level,
		Depth:   baseDepth,
	}}
	for _, child := range block.Children {
		items = append(items, flattenBlock(child, baseDepth+1)...)
	}
	return items
}

func levelFromSubType(subType string) int {
	switch subType {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 0
	}
}

// SearchHistoryOptions 搜索文档历史选项
type SearchHistoryOptions struct {
	Query    string `json:"query"`
	Notebook string `json:"notebook,omitempty"`
	Page     int    `json:"page,omitempty"`
}

// SearchHistoryResponse 搜索历史响应
type SearchHistoryResponse struct {
	Histories  []string `json:"histories"`  // created 时间戳列表
	PageCount  int      `json:"pageCount"`
	TotalCount int      `json:"totalCount"`
}

// SearchHistory 搜索文档历史（第一阶段：获取 created 列表）
// 路由：POST /api/history/searchHistory
func (c *Client) SearchHistory(ctx context.Context, opts SearchHistoryOptions) (*SearchHistoryResponse, error) {
	var ret SearchHistoryResponse
	if err := c.post(ctx, "/api/history/searchHistory", opts, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// GetHistoryItemsOptions 获取历史条目详情选项
type GetHistoryItemsOptions struct {
	Created string `json:"created"`
	Query   string `json:"query,omitempty"`
	Notebook string `json:"notebook,omitempty"`
}

// HistoryItem 历史条目
type HistoryItem struct {
	Title    string `json:"title"`
	Path     string `json:"path"`
	Op       string `json:"op"`
	Notebook string `json:"notebook"`
}

// GetHistoryItems 获取指定时间点的历史条目详情（第二阶段）
// 路由：POST /api/history/getHistoryItems
func (c *Client) GetHistoryItems(ctx context.Context, opts GetHistoryItemsOptions) ([]HistoryItem, error) {
	var ret struct {
		Items []HistoryItem `json:"items"`
	}
	if err := c.post(ctx, "/api/history/getHistoryItems", opts, &ret); err != nil {
		return nil, err
	}
	return ret.Items, nil
}

// RollbackDocHistoryOptions 回滚文档历史选项
type RollbackDocHistoryOptions struct {
	Notebook    string `json:"notebook"`
	HistoryPath string `json:"historyPath"`
}
